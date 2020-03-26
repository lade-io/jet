package pack

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type PythonPack struct {
	WorkDir string
}

func (p *PythonPack) Detect() bool {
	return fileExists(p.WorkDir, "requirements.txt") ||
		fileExists(p.WorkDir, "setup.py") ||
		fileExists(p.WorkDir, "environment.yml") ||
		fileExists(p.WorkDir, "Pipfile")
}

func (p *PythonPack) Metadata() *Metadata {
	user := "web"
	meta := &Metadata{
		Env: map[string]string{
			"PATH":     "/home/" + user + "/.local/bin:$PATH",
			"PIP_USER": "true",
		},
		User: user,
	}
	requirements := p.requirements()
	if !requirements["gunicorn"] {
		meta.Tools = append(meta.Tools, &Tool{
			Name:    "pip",
			Install: []string{"install gunicorn"},
		})
	}
	if requirements["pylibmc"] {
		meta.Packages = append(meta.Packages, "libmemcached-dev")
	}
	switch {
	case fileExists(p.WorkDir, "requirements.txt"):
	case fileExists(p.WorkDir, "setup.py"):
		meta.Tools = append(meta.Tools, &Tool{
			Name:     "pip-compile",
			Download: "pip install pip-tools",
			Files:    []string{"setup.py"},
			Install:  []string{"setup.py"},
		})
	case fileExists(p.WorkDir, "environment.yml"):
		convert := []string{
			"r environment.yml dependencies[*]",
			"sed 's/=/==/;/^python=/d' > requirements.txt",
		}
		meta.Tools = append(meta.Tools, &Tool{
			Name:    "yq",
			Owner:   "mikefarah",
			Files:   []string{"environment.yml"},
			Install: []string{strings.Join(convert, " | ")},
		})
	case fileExists(p.WorkDir, "Pipfile"):
		meta.Tools = append(meta.Tools, &Tool{
			Name:     "pipenv",
			Download: "pip install pipenv",
			Files:    []string{"Pipfile", "Pipfile.lock"},
			Install:  []string{"lock -r > requirements.txt"},
		})
	}
	meta.Tools = append(meta.Tools, &Tool{
		Name:    "pip",
		Files:   []string{"requirements.txt"},
		Install: []string{"install -r requirements.txt"},
	})
	return meta
}

func (p *PythonPack) Name() string {
	return "python"
}

func (p *PythonPack) Command() (string, error) {
	paths, err := fileGlob(p.WorkDir, "**/*.py")
	if err != nil {
		return "", err
	}

	for _, path := range paths {
		file, err := os.Open(filepath.Join(p.WorkDir, path))
		if err != nil {
			continue
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			matches := pyappRegex.FindStringSubmatch(line)
			if len(matches) > 1 {
				path = strings.TrimSuffix(path, filepath.Ext(path))
				path = strings.Replace(path, "/", ".", -1)
				return "gunicorn " + path + ":" + matches[1], nil
			}
		}
	}
	return "", nil
}

func (p *PythonPack) Version() (string, error) {
	names := []string{"runtime.txt", "environment.yml", "Pipfile"}
	for _, name := range names {
		file, err := os.Open(filepath.Join(p.WorkDir, name))
		if err != nil {
			continue
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			matches := pythonRegex.FindStringSubmatch(line)
			if len(matches) > 2 {
				return matches[2], nil
			}
		}
	}
	return "", nil
}

func (p *PythonPack) requirements() map[string]bool {
	requirements := map[string]bool{}
	names := []string{"requirements.txt", "setup.py", "environment.yml", "Pipfile"}
	for _, name := range names {
		file, err := os.Open(filepath.Join(p.WorkDir, name))
		if err != nil {
			continue
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			matches := pipRegex.FindStringSubmatch(line)
			if len(matches) > 1 {
				requirements[matches[1]] = true
			}
		}
		break
	}
	return requirements
}

var (
	pipRegex    = regexp.MustCompile(`(?:^|[-'"])\s*([a-zA-Z_][-\w]*)\s*(?:$|[=<>'"[])`)
	pyappRegex  = regexp.MustCompile(`(\w+)\s*=\s*[\w.]*(Flask|get_wsgi_application)\(`)
	pythonRegex = regexp.MustCompile(`^-?\s*python(_version)?[-=\s'"]*([.x*\d]+)['"]?$`)
)
