package pack

type GoPack struct {
	WorkDir string
}

func (g *GoPack) Detect() bool {
	return fileExists(g.WorkDir, "Gopkg.toml") ||
		fileExists(g.WorkDir, "glide.yaml") ||
		fileExists(g.WorkDir, "Godeps/Godeps.json") ||
		fileExists(g.WorkDir, "go.mod") ||
		fileExists(g.WorkDir, "vendor/vendor.json")
}

func (g *GoPack) Metadata() *Metadata {
	meta := &Metadata{
		Install: []string{"go install -v -ldflags '-s -w' ."},
	}
	switch {
	case fileExists(g.WorkDir, "Gopkg.toml"):
		meta.Tools = append(meta.Tools, &Tool{
			Name:    "dep",
			Owner:   "golang",
			Files:   []string{"Gopkg.toml", "Gopkg.lock"},
			Install: []string{"ensure -vendor-only"},
		})
	case fileExists(g.WorkDir, "glide.yaml"):
		meta.Root = Root{
			File: "glide.yaml",
			Key:  "package",
		}
		meta.Tools = append(meta.Tools, &Tool{
			Name:    "glide",
			Owner:   "Masterminds",
			Files:   []string{"glide.yaml", "glide.lock"},
			Install: []string{"install -v", "cache-clear"},
		})
	case fileExists(g.WorkDir, "Godeps/Godeps.json"):
		meta.Root = Root{
			File: "Godeps/Godeps.json",
			Key:  "ImportPath",
		}
		meta.Tools = append(meta.Tools, &Tool{
			Name:    "godep",
			Owner:   "tools",
			Files:   []string{"Godeps/Godeps.json"},
			Install: []string{"restore"},
		})
	case fileExists(g.WorkDir, "go.mod"):
		meta.Env = map[string]string{"GO111MODULE": "on"}
		meta.Root = Root{
			File: "go.mod",
			Key:  "module",
		}
		meta.Tools = append(meta.Tools, &Tool{
			Name:    "go mod",
			Files:   []string{"go.mod", "go.sum"},
			Install: []string{"download"},
		})
	case fileExists(g.WorkDir, "vendor/vendor.json"):
		meta.Root = Root{
			File: "vendor/vendor.json",
			Key:  "rootPath",
		}
		meta.Tools = append(meta.Tools, &Tool{
			Name:    "govendor",
			Owner:   "kardianos",
			Files:   []string{"vendor/vendor.json"},
			Install: []string{"sync"},
		})
	}
	return meta
}

func (g *GoPack) Name() string {
	return "golang"
}

func (g *GoPack) Command() (string, error) {
	return "", nil
}

func (g *GoPack) Version() (string, error) {
	return "", nil
}
