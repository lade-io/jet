package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "jet",
	Short: "Convert source code into Docker images",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		cmd.SilenceUsage = true
	},
}

func SetVersion(version string) {
	RootCmd.Version = version
}

func init() {
	cobra.OnInitialize(initLogger)
	RootCmd.SetHelpTemplate(helpTemplate)
	RootCmd.SetUsageTemplate(usageTemplate)

	RootCmd.PersistentFlags().BoolP("help", "h", false, "Print help message")
	RootCmd.Flags().BoolP("version", "v", false, "Print version and exit")

	RootCmd.AddCommand(buildCmd)
	RootCmd.AddCommand(debugCmd)
	RootCmd.AddCommand(versionCmd)
}

func initLogger() {
	log.SetFlags(0)
}
