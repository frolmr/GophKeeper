// list implements the records list command.
//
// This handles:
// - List all users records
package record

import (
	"fmt"
	"os"

	"github.com/frolmr/GophKeeper/internal/client/app"
	"github.com/frolmr/GophKeeper/internal/client/app/record"
	"github.com/jedib0t/go-pretty/table"
	"github.com/spf13/cobra"
)

// NewListCommand creates the 'record' list command.
//
// Returns:
//
//	*cobra.Command: The configured list subcommand
func NewListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List users records",
		Long:  "Interactive command to list all users records",
		Run: func(cmd *cobra.Command, args []string) {
			needAuth := true
			needCrypto := false

			app, err := app.NewApp(needAuth, needCrypto)
			if err != nil {
				fmt.Printf("Failed to setup cli: %v", err)
			}
			defer app.Close()

			service := record.NewRecordService(app.Storage, app.CryptoService, app.ConnAdapter)

			records, err := service.ListRecords()
			if err != nil {
				fmt.Printf("failed to list users devices: %v", err)
				return
			}

			t := table.NewWriter()
			t.SetOutputMirror(os.Stdout)
			t.AppendHeader(table.Row{"UUID", "Name", "Kind"})
			for _, record := range records {
				t.AppendRows([]table.Row{{record.UUID, record.Name, record.Kind}})
			}
			t.Render()
		},
	}
	return cmd
}
