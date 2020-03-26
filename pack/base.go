package pack

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"text/template"

	"github.com/docker/libcompose/docker/builder"
	"github.com/docker/libcompose/docker/client"
	"github.com/docker/libcompose/logger"
)

type Pack interface {
	Detect() bool
	Metadata() *Metadata
	Name() string
	Command() (string, error)
	Version() (string, error)
}

type Metadata struct {
	Command  string
	Depends  []*Depend
	Env      map[string]string
	Install  []string
	Name     string
	Packages []string
	Path     string
	Process  []string
	Root     Root
	Sources  []*Source
	Tools    []*Tool
	User     string
	Variant  string
	Version  string
}

type Depend struct {
	Name string
	Args []string
	List bool
}

type Root struct {
	File string
	Key  string
}

type Source struct {
	Entry string
	File  string
	Key   string
}

type Tool struct {
	Copy     map[string][]string
	Name     string
	Owner    string
	Archive  bool
	Binary   bool
	Download string
	Files    []string
	Install  []string
	Hook     func(meta *Metadata, tool *Tool) error
}

func Detect(workDir string) (pack *Buildpack, err error) {
	packs := []Pack{
		&GoPack{workDir},
		&PhpPack{workDir},
		&PythonPack{workDir},
		&RubyPack{workDir},
		&NodePack{workDir},
	}

	detected := []*Buildpack{}
	for _, p := range packs {
		if !p.Detect() {
			continue
		}

		pack = &Buildpack{}
		pack.Metadata = p.Metadata()
		pack.Metadata.Name = p.Name()
		pack.Metadata.Command, err = p.Command()
		if err != nil {
			return nil, err
		}

		pack.Metadata.Version, err = p.Version()
		if err != nil {
			return nil, err
		}

		pack.WorkDir = workDir
		detected = append(detected, pack)
	}

	if len(detected) < 1 {
		return nil, errors.New("No known buildpacks support this app")
	}

	pack = detected[0]
	err = getPath(workDir, pack.Metadata)
	if err != nil {
		return nil, err
	}

	err = getVersion(pack.Metadata)
	if err != nil {
		return nil, err
	}

	err = getTools(workDir, pack.Metadata)
	return
}

type Buildpack struct {
	Metadata *Metadata
	WorkDir  string
}

func (b *Buildpack) BuildImage(name string) (string, error) {
	if err := b.createDockerfile(); err != nil {
		return "", err
	}

	if err := b.createDockerignore(); err != nil {
		return "", err
	}

	imageClient, err := client.Create(client.Options{})
	if err != nil {
		return "", err
	}
	defer imageClient.Close()

	logger := &buildLogger{}
	daemon := builder.DaemonBuilder{
		Client:           imageClient,
		ContextDirectory: b.WorkDir,
		Dockerfile:       ".jet/Dockerfile",
		LoggerFactory:    logger,
	}

	ctx := context.Background()
	return logger.imageID, daemon.Build(ctx, name)
}

func (b *Buildpack) GetDockerfile() (string, error) {
	out := &bytes.Buffer{}
	if err := dockerTemplate.Execute(out, b.Metadata); err != nil {
		return "", err
	}
	return out.String(), nil
}

func (b *Buildpack) createDockerfile() error {
	dockerfile, err := b.GetDockerfile()
	if err != nil {
		return err
	}

	err = os.MkdirAll(filepath.Join(b.WorkDir, ".jet"), 0755)
	if err != nil {
		return err
	}

	tempFile := filepath.Join(b.WorkDir, ".jet/Dockerfile")
	return ioutil.WriteFile(tempFile, []byte(dockerfile), 0644)
}

func (b *Buildpack) createDockerignore() error {
	files := []string{".jet", ".dockerignore"}
	diff := map[string]bool{}
	for _, f := range files {
		diff[f] = true
	}

	dockerignore := filepath.Join(b.WorkDir, ".dockerignore")
	file, err := os.OpenFile(dockerignore, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if diff[line] {
			diff[line] = false
		}
	}

	w := bufio.NewWriter(file)
	for line, keep := range diff {
		if !keep {
			continue
		}

		_, err := w.WriteString(line + "\n")
		if err != nil {
			return err
		}
	}
	return w.Flush()
}

var (
	dockerTemplate = template.Must(template.New("Dockerfile").Parse(dockerString))
	imageRegex     = regexp.MustCompile(`^ ---> ([0-9a-f]+)\s*$`)
)

type buildLogger struct {
	logger.RawLogger
	imageID string
}

func (b *buildLogger) Out(message []byte) {
	msg := string(message)
	matches := imageRegex.FindStringSubmatch(msg)
	if len(matches) > 1 {
		b.imageID = matches[1]
	}
	fmt.Print(msg)
}

func (b *buildLogger) CreateBuildLogger(_ string) logger.Logger {
	return b
}
