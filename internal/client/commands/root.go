package commands

import (
	"os"

	"github.com/frolmr/GophKeeper/internal/client/commands/device"
	"github.com/frolmr/GophKeeper/internal/client/commands/record"
	"github.com/frolmr/GophKeeper/internal/client/commands/user"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "GophKeeper",
	Short: "CLI tool for super secret info storage",
	Long: `GophKeeper is CLI tool that manages private data securely
Usage:
  GophKeeper COMMAND [SUBCOMMAND] [FLAGS] [ARGUMENTS]

Available Commands:
  user       User account operations
    register  - Create a new user account
    login     - Authenticate and get access token

  device     Device management
    add       - Register a new device
    confirm   - Confirm device registration
    list      - List all registered devices

  record     Secure record management
    add       - Add a new secret record
    read      - View a secret record
    update    - Modify an existing record
    delete    - Remove a record
    list      - List all records

  version    Show application version

Examples:
  GophKeeper user register
  GophKeeper device add
  GophKeeper record add credentials
`,
}

func Execute() {
	registerCommands()
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func registerCommands() {
	rootCmd.AddCommand(user.NewUserCommand())
	rootCmd.AddCommand(device.NewDeviceCommand())
	rootCmd.AddCommand(record.NewRecordCommand())
}
