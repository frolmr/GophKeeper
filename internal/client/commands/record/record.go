// Package record provides commands for user records management.
//
// This includes:
// - New record
// - Read record
// - Update record
// - List records
// - Delete record
//
// All commands in this package interact with the server's user API.
package record

import (
	"github.com/spf13/cobra"
)

// NewRecordCommand creates the root 'record' command with all its subcommands.
//
// Returns:
//
//	*cobra.Command: The configured record command with add, read, update, delete and list subcommands
func NewRecordCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "record",
		Short: "Manage records",
	}

	cmd.AddCommand(
		NewAddCommand(),
		NewReadCommand(),
		NewUpdateCommand(),
		NewListCommand(),
		NewDeleteCommand(),
	)

	return cmd
}
