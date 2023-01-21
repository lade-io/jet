package pack

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type RubyPack struct {
	WorkDir string
}

func (r *RubyPack) Detect() bool {
	return fileExists(r.WorkDir, "Gemfile")
}

func (r *RubyPack) Metadata() *Metadata {
	meta := &Metadata{
		User: "web",
	}
	specs := r.specs()
	if _, ok := specs["puma"]; !ok {
		meta.Tools = append(meta.Tools, &Tool{
			Name:    "gem",
			Install: []string{"install puma"},
		})
	}

	meta.Tools = append(meta.Tools, &Tool{
		Name:    "bundle",
		Files:   []string{"Gemfile", "Gemfile.lock"},
		Install: []string{"install"},
	})

	yarn := fileExists(r.WorkDir, "yarn.lock")
	node := yarn
	for _, name := range []string{"execjs", "webpacker"} {
		if _, ok := specs[name]; ok {
			node = true
			break
		}
	}

	if node {
		meta.Tools = append(meta.Tools, &Tool{
			Name:  "node",
			Owner: "nodejs",
		})
	}

	if yarn {
		meta.Tools = append(meta.Tools, &Tool{
			Name:     "yarn",
			Download: "corepack enable",
			Files:    []string{"package.json", "yarn.lock"},
			Install:  []string{"install"},
		})
	}
	return meta
}

func (r *RubyPack) Name() string {
	engine, _, _ := r.version()
	if engine != "" {
		return engine
	}
	return "ruby"
}

func (r *RubyPack) Command() (string, error) {
	if fileExists(r.WorkDir, "config.ru") {
		return "puma -p ${PORT-3000}", nil
	}
	return "", nil
}

func (r *RubyPack) Version() (string, error) {
	engine, engine_version, ruby_version := r.version()
	if engine != "" && engine_version != "" {
		return engine_version, nil
	}
	return ruby_version, nil
}

func (r *RubyPack) specs() map[string]string {
	file, err := os.Open(filepath.Join(r.WorkDir, "Gemfile.lock"))
	if err != nil {
		return nil
	}
	defer file.Close()

	specs := map[string]string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		matches := specRegex.FindStringSubmatch(line)
		if len(matches) > 2 {
			specs[matches[1]] = matches[2]
		}
	}
	return specs
}

func (r *RubyPack) version() (string, string, string) {
	file, err := os.Open(filepath.Join(r.WorkDir, "Gemfile"))
	if err != nil {
		return "", "", ""
	}
	defer file.Close()

	version := map[string]string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "ruby") {
			continue
		}

		matches := rubyRegex.FindAllStringSubmatch(line, -1)
		replacer := strings.NewReplacer("\"", "", "'", "")
		for _, match := range matches {
			if len(match) > 2 {
				version[match[1]] = replacer.Replace(match[2])
			}
		}
		break
	}
	return version["engine"], version["engine_version"], version["ruby"]
}

var (
	specRegex = regexp.MustCompile(`^\s+([-\w]+)\s\(([.\d]+)\)`)
	rubyRegex = regexp.MustCompile(`['"]?([a-z_]+)['"]?[:=>\s]*['"]([a-z]*[^a-z]*\w)['"]`)
)
