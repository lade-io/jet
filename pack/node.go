package pack

import (
	"encoding/json"
	"strings"
)

type NodePack struct {
	WorkDir string
}

func (n *NodePack) Detect() bool {
	return fileExists(n.WorkDir, "package.json")
}

func (n *NodePack) Metadata() *Metadata {
	meta := &Metadata{}
	if fileExists(n.WorkDir, "yarn.lock") {
		meta.Packages = append(meta.Packages, "yarn")
		meta.Sources = append(meta.Sources, &Source{
			Entry: "deb http://dl.yarnpkg.com/debian/ stable main",
			Key:   "https://dl.yarnpkg.com/debian/pubkey.gpg",
			File:  "yarn.list",
		})
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
		})
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

		node = strings.Replace(node, "~>", "~", -1)
		return node, nil
	}
	return "", nil
}
