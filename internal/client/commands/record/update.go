// update implements the record update command.
//
// This handles:
// - Record update
package record

import (
	"fmt"
	"os"

	"github.com/frolmr/GophKeeper/internal/client/app"
	"github.com/frolmr/GophKeeper/internal/client/app/record"
	"github.com/frolmr/GophKeeper/internal/client/input"
	"github.com/spf13/cobra"
)

// NewUpdateCommand creates the 'record' udpate command.
//
// Returns:
//
//	*cobra.Command: The configured update subcommand
func NewUpdateCommand() *cobra.Command {
	var name, kind, id string

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update record",
		Long:  "Interactive command to update existing users sercret record",
		Run: func(cmd *cobra.Command, args []string) {
			if name == "" || kind == "" || id == "" {
				fmt.Println("record name, kind and ID are required")
				return
			}

			inReader := input.NewRecordInputReader(os.Stdin)
			rec, err := inReader.ReadInput(name, kind)
			if err != nil {
				fmt.Printf("error reading input: %v", err)
				return
			}

			needAuth := true
			needCrypto := true

			app, err := app.NewApp(needAuth, needCrypto)
			if err != nil {
				fmt.Printf("Failed to setup cli: %v", err)
			}
			defer app.Close()

			service := record.NewRecordService(app.Storage, app.CryptoService, app.ConnAdapter)

			if err := service.UpdateRecord(rec, id); err != nil {
				fmt.Printf("Update record command failed: %v", err.Error())
			} else {
				fmt.Println("Record updated successful!")
			}
		},
	}
	cmd.Flags().StringVarP(&name, "name", "n", "", "record name")
	cmd.Flags().StringVarP(&kind, "kind", "k", "", "record kind")
	cmd.Flags().StringVarP(&id, "record_id", "i", "", "record UUID")

	return cmd
}
