package commands

import (
	"os"

	"github.com/frolmr/GophKeeper/internal/client/app"
	"github.com/spf13/cobra"
)

var gk *app.GophKeeper

func SetApp(gkApp *app.GophKeeper) {
	gk = gkApp
}

var rootCmd = &cobra.Command{
	Use:   "GophKeeper",
	Short: "CLI tool for super secret info storage",
	Long: `GophKeeper is CLI tool that manages private data securely
Usage: GophKeeper COMMAND [INPUT]

Available Commands:
  register      Register new user
`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
