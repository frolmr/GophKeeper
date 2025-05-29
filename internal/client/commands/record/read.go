// read implements the record read command.
//
// This handles:
// - Record reading
package record

import (
	"fmt"

	"github.com/frolmr/GophKeeper/internal/client/app"
	"github.com/frolmr/GophKeeper/internal/client/app/record"
	"github.com/spf13/cobra"
)

// NewReadCommand creates the 'record' read command.
//
// Returns:
//
//	*cobra.Command: The configured read subcommand
func NewReadCommand() *cobra.Command {
	var id string

	cmd := &cobra.Command{
		Use:   "read",
		Short: "Read users record",
		Long:  "Interactive command to read users record",
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

			rec, err := service.ReadRecord(id)
			if err != nil {
				fmt.Printf("Record deletion failed: %v", err.Error())
				return
			}

			fmt.Println(rec)
		},
	}

	cmd.Flags().StringVarP(&id, "record_id", "i", "", "record UUID")

	return cmd
}
