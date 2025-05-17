package commands

import (
	"fmt"

	"github.com/frolmr/GophKeeper/pkg/buildinfo"
	"github.com/spf13/cobra"
)

//nolint:gochecknoinits // need for command module
func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Version and build date",
	Long:  "Print the version number and build date of GophKeeper",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Version:    %s\n", buildinfo.CurrentBuild.Version)
		fmt.Printf("Build Date: %s\n", buildinfo.CurrentBuild.BuildDate)
	},
}
