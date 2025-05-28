// Package user provides commands for user account management.
//
// This includes:
// - User registration
// - User authentication
//
// All commands in this package interact with the server's user API.
package user

import (
	"github.com/spf13/cobra"
)

// NewUserCommand creates the root 'user' command with all its subcommands.
//
// Returns:
//
//	*cobra.Command: The configured user command with register and login subcommands
func NewUserCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user",
		Short: "Manage user",
	}

	cmd.AddCommand(
		NewLoginCommand(),
		NewRegisterCommand(),
	)

	return cmd
}
