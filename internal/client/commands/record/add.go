// add implements the record add command.
//
// This handles:
// - New record creation
package record

import (
	"fmt"
	"os"

	"github.com/frolmr/GophKeeper/internal/client/app"
	"github.com/frolmr/GophKeeper/internal/client/app/record"
	"github.com/frolmr/GophKeeper/internal/client/input"
	"github.com/spf13/cobra"
)

// NewAddCommand creates the 'record' add command.
//
// Returns:
//
//	*cobra.Command: The configured add subcommand
func NewAddCommand() *cobra.Command {
	var name, kind string

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add new record",
		Long:  "Interactive command to add new users sercret record",
		Run: func(cmd *cobra.Command, args []string) {
			if name == "" || kind == "" {
				fmt.Println("record name and kind are required")
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

			if err := service.AddRecord(rec); err != nil {
				fmt.Printf("Add record command failed: %v", err.Error())
			} else {
				fmt.Println("Record added successful!")
			}
		},
	}
	cmd.Flags().StringVarP(&name, "name", "n", "", "record name")
	cmd.Flags().StringVarP(&kind, "kind", "k", "", "record kind")

	return cmd
}
