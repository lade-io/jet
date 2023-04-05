package pack

import (
	"encoding/json"
	"strings"

	"github.com/aquasecurity/go-version/pkg/version"
)

type NodePack struct {
	WorkDir string
}

func (n *NodePack) Detect() bool {
	return fileExists(n.WorkDir, "package.json")
}

func (n *NodePack) Metadata() *Metadata {
	user := "node"
	meta := &Metadata{
		Env: map[string]string{
			"PATH": "/home/" + user + "/app/node_modules/.bin:$PATH",
		},
		User: user,
	}

	if fileExists(n.WorkDir, "yarn.lock") {
		meta.Tools = append(meta.Tools, &Tool{
			Name:    "yarn",
			Files:   []string{"**/package.json", "yarn.lock"},
			Install: []string{"install"},
		})
	} else {
		meta.Tools = append(meta.Tools, &Tool{
			Name:    "npm",
			Files:   []string{"package.json", "package-lock.json"},
			Install: []string{"install"},
			Hook: func(meta *Metadata, tool *Tool) error {
				if !fileExists(n.WorkDir, "package-lock.json") {
					return nil
				}

				constraints, err := version.NewConstraints("^8.12 || >=10.3")
				if err != nil {
					return err
				}

				v, _ := version.Parse(meta.Version)
				if constraints.Check(v) {
					tool.Install = []string{"ci"}
				}
				return nil
			},
		})
	}

	scripts := n.scripts()
	if scripts["build"] {
		meta.Install = append(meta.Install, meta.Tools[0].Name+" run build")
	}
	return meta
}

func (n *NodePack) Name() string {
	return "node"
}

func (n *NodePack) Command() (string, error) {
	b, err := fileRead(n.WorkDir, "package.json")
	if err != nil {
		return "", err
	}

	conf := map[string]interface{}{}
	if err = json.Unmarshal(b, &conf); err != nil {
		return "", err
	}

	main, ok := conf["main"].(string)
	if ok {
		return "node " + main, nil
	}

	scripts, ok := conf["scripts"].(map[string]interface{})
	if !ok {
		return "", nil
	}

	start, ok := scripts["start"].(string)
	if ok {
		return start, nil
	}
	return "", nil
}

func (n *NodePack) Version() (string, error) {
	for _, file := range []string{"package.json", ".nvmrc"} {
		b, err := fileRead(n.WorkDir, file)
		if err != nil {
			continue
		}

		if file == ".nvmrc" {
			return string(b), nil
		}

		conf := map[string]interface{}{}
		if err = json.Unmarshal(b, &conf); err != nil {
			return "", err
		}

		engines, ok := conf["engines"].(map[string]interface{})
		if !ok {
			continue
		}

		node, ok := engines["node"].(string)
		if !ok {
			continue
		}

		node = strings.ReplaceAll(node, "~>", "~")
		return node, nil
	}
	return "", nil
}

func (n *NodePack) scripts() map[string]bool {
	b, err := fileRead(n.WorkDir, "package.json")
	if err != nil {
		return nil
	}

	conf := map[string]interface{}{}
	if err = json.Unmarshal(b, &conf); err != nil {
		return nil
	}

	scripts, ok := conf["scripts"].(map[string]interface{})
	if !ok {
		return nil
	}

	scriptMap := map[string]bool{}
	for key, value := range scripts {
		if _, ok := value.(string); ok {
			scriptMap[key] = true
		}
	}
	return scriptMap
}
