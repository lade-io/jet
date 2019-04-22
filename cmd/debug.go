package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/lade-io/jet/pack"
	"github.com/spf13/cobra"
)

var debugCmd = &cobra.Command{
	Use:   "debug <path>",
	Short: "Print generated Dockerfile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		workDir, err := filepath.Abs(args[0])
		if err != nil {
			return err
		}
		return debugRun(workDir)
	},
}

func debugRun(workDir string) error {
	bp, err := pack.Detect(workDir)
	if err != nil {
		return err
	}

	dockerfile, err := bp.GetDockerfile()
	if err != nil {
		return err
	}

	fmt.Print(dockerfile)
	return nil
}
