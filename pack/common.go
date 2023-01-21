package pack

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/bmatcuk/doublestar"
	"github.com/cloudingcity/gomod"
	"github.com/docker/distribution/reference"
	"github.com/docker/distribution/registry/client"
	"github.com/google/go-github/v45/github"
	"github.com/gregjones/httpcache"
	"github.com/gregjones/httpcache/diskcache"
	"github.com/hashicorp/go-version"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v3"
)

var (
	cacheExpiry = time.Hour
	httpClient  *http.Client
	transport   *cacheTransport
)

type cacheTransport struct {
	maxAge int
	rt     http.RoundTripper
}

type nodeVersion struct {
	Version string      `json:"version"`
	LTS     interface{} `json:"lts"`
}

func (c *cacheTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	req.Header.Set("Cache-Control", fmt.Sprintf("max-age=%d", c.maxAge))
	resp, err = c.rt.RoundTrip(req)
	if resp.StatusCode != http.StatusOK {
		resp.Header.Set("Cache-Control", "no-cache")
	}
	return
}

func (n *nodeVersion) Download() string {
	return fmt.Sprintf("https://nodejs.org/dist/%[1]s/node-%[1]s-linux-x64.tar.gz", n.Version)
}

func (n *nodeVersion) IsLTS() bool {
	switch data := n.LTS.(type) {
	case bool:
		return data
	case string:
		return true
	}
	return false
}

func init() {
	cache := diskcache.New(filepath.Join(os.TempDir(), "jet-cache"))
	transport = &cacheTransport{
		maxAge: int(cacheExpiry / time.Second),
		rt:     httpcache.NewTransport(cache),
	}
	if accessToken, exists := os.LookupEnv("GITHUB_TOKEN"); exists {
		token := &oauth2.Token{AccessToken: accessToken}
		tokenSource := oauth2.StaticTokenSource(token)
		httpClient = &http.Client{
			Transport: &oauth2.Transport{
				Base:   transport,
				Source: tokenSource,
			},
		}
	} else {
		httpClient = &http.Client{Transport: transport}
	}
}

func getDownload(tool *Tool) (err error) {
	name := tool.Name
	owner := tool.Owner
	if owner == "" {
		return nil
	}

	defer func() {
		if err == nil && tool.Download == "" {
			err = fmt.Errorf("%s tool not found", name)
		}
	}()

	if name == "node" {
		tool.Archive = "/usr/local"
		tool.Download, err = getNodeDownload()
		return err
	}

	client := github.NewClient(httpClient)
	ctx := context.Background()
	release, _, err := client.Repositories.GetLatestRelease(ctx, owner, name)
	if err != nil {
		return err
	}

	binary, err := regexp.Compile(name + `(.*linux[-_]amd64($|\.tar\.gz)|\.phar)`)
	if err != nil {
		return err
	}

	for _, asset := range release.Assets {
		if binary.MatchString(asset.GetName()) {
			tool.Download = asset.GetBrowserDownloadURL()
			if strings.HasSuffix(tool.Download, ".tar.gz") {
				tool.Archive = "/usr/local/bin"
			} else {
				tool.Binary = true
			}
			break
		}
	}
	return nil
}

func getNodeDownload() (string, error) {
	resp, err := httpClient.Get("https://nodejs.org/dist/index.json")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var versions []nodeVersion
	if err = json.NewDecoder(resp.Body).Decode(&versions); err != nil {
		return "", err
	}

	for _, version := range versions {
		if version.IsLTS() {
			return version.Download(), nil
		}
	}
	return "", nil
}

func getPath(dir string, meta *Metadata) error {
	defer func() {
		current := filepath.Clean(meta.Path) == "."
		if current {
			meta.Path = "/app"
		}
		if meta.Name == "golang" {
			meta.Path = filepath.Join("/go/src", meta.Path)
			meta.Command = filepath.Base(meta.Path)
		} else if current {
			meta.Path = filepath.Join("/home", meta.User, meta.Path)
		}
		meta.Path += "/"
		getProcess(meta)
	}()

	meta.Tools = append(meta.Tools, &Tool{
		Files:   []string{"."},
		Install: meta.Install,
	})

	file := meta.Root.File
	if file == "" {
		return nil
	}

	b, err := fileRead(dir, file)
	if err != nil {
		return err
	}

	conf := map[string]interface{}{}
	switch strings.ToLower(filepath.Ext(file)) {
	case ".json":
		err = json.Unmarshal(b, &conf)
	case ".yaml":
		err = yaml.Unmarshal(b, &conf)
	case ".mod":
		conf["module"], err = getModule(b)
	}
	if err != nil {
		return err
	}

	if path, ok := conf[meta.Root.Key].(string); ok {
		meta.Path = filepath.Clean(path)
	}
	return nil
}

func getModule(data []byte) (string, error) {
	mod, err := gomod.Parse(data)
	if err != nil {
		return "", err
	}
	return mod.Module.Path, nil
}

func getProcess(meta *Metadata) {
	if strings.Contains(meta.Command, "$") {
		meta.Process = []string{"sh", "-c", meta.Command}
	} else if meta.Command != "" {
		meta.Process = strings.Split(meta.Command, " ")
	}
}

func getTags(name string) ([]string, error) {
	namedRef, err := reference.WithName("library/" + name)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	repository, err := client.NewRepository(namedRef, "https://hub.lade.io", transport)
	if err != nil {
		return nil, err
	}

	return repository.Tags(ctx).All(ctx)
}

func getTools(dir string, meta *Metadata) error {
	for _, tool := range meta.Tools {
		if tool.Hook != nil {
			if err := tool.Hook(meta, tool); err != nil {
				return err
			}
		}

		copyMap, err := fileCopy(dir, tool.Files)
		if err != nil {
			return err
		}

		tool.Copy = copyMap
		if err = getDownload(tool); err != nil {
			return err
		}
	}
	return nil
}

func getVersion(meta *Metadata) error {
	tags, err := getTags(meta.Name)
	if err != nil {
		return err
	}

	versions := []string{}
	for _, tag := range tags {
		var v *version.Version
		v, err = version.NewVersion(tag)
		if err != nil || v.Prerelease() != meta.Variant {
			continue
		}
		versions = append(versions, tag)
	}

	sort.Slice(versions, func(i, j int) bool {
		v1, _ := version.NewVersion(versions[i])
		v2, _ := version.NewVersion(versions[j])
		if v1.Equal(v2) {
			return versions[i] > versions[j]
		}
		return v1.GreaterThan(v2)
	})

	meta.Version = strings.TrimRight(meta.Version, ".x*")
	if meta.Version == "" {
		meta.Version = ">0"
	}

	constraints, err := version.NewConstraint(meta.Version)
	if err != nil {
		return err
	}

	for _, tag := range versions {
		ver := strings.Split(tag, "-")[0]
		if meta.Version < ver {
			continue
		}

		v, _ := version.NewVersion(ver)
		if constraints.Check(v) {
			meta.Version = tag
			return nil
		}
	}
	return fmt.Errorf("Unknown %s version %s", meta.Name, meta.Version)
}

func fileCopy(dir string, files []string) (map[string][]string, error) {
	paths := []string{}
	for _, file := range files {
		glob, err := fileGlob(dir, file)
		if err != nil {
			return nil, err
		}
		paths = append(paths, glob...)
	}

	copyMap := map[string][]string{}
	for _, path := range paths {
		dest := filepath.Dir(path)
		copyMap[dest] = append(copyMap[dest], path)
	}
	return copyMap, nil
}

func fileExists(dir, file string) bool {
	_, err := os.Stat(filepath.Join(dir, file))
	return err == nil
}

func fileGlob(dir, file string) ([]string, error) {
	paths, err := doublestar.Glob(filepath.Join(dir, file))
	if err != nil {
		return nil, err
	}

	glob := []string{}
	for _, path := range paths {
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			continue
		}
		glob = append(glob, rel)
	}
	return glob, nil
}

func fileRead(dir, file string) ([]byte, error) {
	return ioutil.ReadFile(filepath.Join(dir, file))
}

const dockerString = `FROM {{.Name}}:{{.Version}}
{{if .Packages}}
RUN set -ex \
	&& apt-get update && apt-get install -y \
{{range .Packages}}		{{.}} \
{{end}}	&& rm -rf /var/lib/apt/lists/*
{{end}}{{range .Tools}}{{if .Download}}{{if .Archive}}
RUN wget -qO {{.Name}}.tar.gz "{{.Download}}" \
	&& tar -xzf {{.Name}}.tar.gz -C {{.Archive}} --strip-components=1 \
	&& rm {{.Name}}.tar.gz
{{else if .Binary}}
RUN wget -qO {{.Name}} "{{.Download}}" \
	&& chmod +x {{.Name}} && mv {{.Name}} /usr/local/bin
{{else}}
RUN {{.Download}}
{{end}}{{end}}{{end}}{{range $d := .Depends}}
RUN {{if $d.List}}{{$d.Name}}{{end}}{{range $i, $e := .Args}}{{if or $d.List $i}} \
	{{if not $d.List}}&& {{end}}{{end}}{{if and (not $d.List) $d.Name}}
{{- $d.Name}} {{end}}{{$e}}{{end}}
{{end}}{{if .Env}}
{{range $key, $val := .Env}}ENV {{$key}}={{$val}}
{{end}}{{end}}{{if eq .User "web"}}
RUN groupadd --gid 1000 {{.User}} \
	&& useradd --uid 1000 --gid {{.User}} --shell /bin/bash --create-home {{.User}}
{{end}}
USER {{.User}}
RUN mkdir -p {{.Path}}
WORKDIR {{.Path}}
{{range $t := .Tools}}{{if or .Copy .Install}}{{range $dir, $files := .Copy}}
COPY {{if $.User}}--chown={{$.User}}:{{$.User}} {{end}}
{{- range $files}}{{.}} {{end}}{{$dir}}/{{end}}{{if .Install}}
RUN {{range $i, $e := .Install}}{{if $i}} \
	&& {{end}}{{if $t.Name}}{{$t.Name}} {{end}}{{$e}}{{end}}{{end}}
{{end}}{{end}}{{if .Process}}
CMD [{{range $i, $e := .Process}}{{if $i}}, {{end}}"{{$e}}"{{end}}]
{{end -}}
`
