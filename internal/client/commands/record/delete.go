// delete implements the record delete command.
//
// This handles:
// - Record deletion
package record

import (
	"fmt"

	"github.com/frolmr/GophKeeper/internal/client/app"
	"github.com/frolmr/GophKeeper/internal/client/app/record"
	"github.com/spf13/cobra"
)

// NewDeleteCommand creates the 'record' delete command.
//
// Returns:
//
//	*cobra.Command: The configured delete subcommand
func NewDeleteCommand() *cobra.Command {
	var id string

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete users record",
		Long:  "Interactive command to confirm unconfirmed users device",
		Run: func(cmd *cobra.Command, args []string) {
			if id == "" {
				fmt.Println("record id is required")
				return
			}

			needAuth := true
			needCrypto := false

			app, err := app.NewApp(needAuth, needCrypto)
			if err != nil {
				fmt.Printf("Failed to setup cli: %v", err)
			}
			defer app.Close()

			service := record.NewRecordService(app.Storage, app.CryptoService, app.ConnAdapter)

			if err := service.DeleteRecord(id); err != nil {
				fmt.Printf("Record deletion failed: %v", err.Error())
			} else {
				fmt.Println("Record deleted")
			}
		},
	}

	cmd.Flags().StringVarP(&id, "record_id", "i", "", "record UUID")

	return cmd
}
