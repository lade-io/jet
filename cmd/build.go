package cmd

import (
	"path/filepath"

	"github.com/lade-io/jet/pack"
	"github.com/spf13/cobra"
)

var buildCmd = func() *cobra.Command {
	var imageName string
	cmd := &cobra.Command{
		Use:   "build <path>",
		Short: "Build a Docker image from source",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			workDir, err := filepath.Abs(args[0])
			if err != nil {
				return err
			}
			if imageName == "" {
				imageName = filepath.Base(workDir)
			}
			return buildRun(imageName, workDir)
		},
	}
	cmd.Flags().StringVarP(&imageName, "name", "n", "", "Image Name")
	return cmd
}()

func buildRun(imageName, workDir string) error {
	bp, err := pack.Detect(workDir)
	if err != nil {
		return err
	}

	_, err = bp.BuildImage(imageName)
	return err
}
