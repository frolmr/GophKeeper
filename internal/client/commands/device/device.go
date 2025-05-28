// Package device provides commands for user device management.
//
// This includes:
// - Addition of the new device
// - List of all user devices
// - Confirmation of user device
//
// All commands in this package interact with the server's user API.

package device

import (
	"github.com/spf13/cobra"
)

// NewDeviceCommand creates the root 'device' command with all its subcommands.
//
// Returns:
//
//	*cobra.Command: The configured device command with add, list and confirm subcommands
func NewDeviceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "device",
		Short: "Manage device",
	}

	cmd.AddCommand(
		NewAddCommand(),
		NewListCommand(),
		NewConfirmCommand(),
	)

	return cmd
}
