// list implements the device list command.
//
// This handles:
// - Listin all users devices

package device

import (
	"fmt"
	"os"

	"github.com/frolmr/GophKeeper/internal/client/app"
	"github.com/frolmr/GophKeeper/internal/client/app/device"
	"github.com/jedib0t/go-pretty/table"
	"github.com/spf13/cobra"
)

// NewListCommand creates the 'device' list command.
//
// Returns:
//
//	*cobra.Command: The configured list subcommand
func NewListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List users devices",
		Long:  "Interactive command to list all users devices",
		Run: func(cmd *cobra.Command, args []string) {
			needAuth := true
			needCrypto := false

			app, err := app.NewApp(needAuth, needCrypto)
			if err != nil {
				fmt.Printf("Failed to setup cli: %v", err)
			}
			defer app.Close()

			service := device.NewDeviceService(app.Storage, app.CryptoService, app.ConnAdapter)

			devices, err := service.ListDevices()
			if err != nil {
				fmt.Printf("failed to list users devices: %v", err)
				return
			}

			t := table.NewWriter()
			t.SetOutputMirror(os.Stdout)
			t.AppendHeader(table.Row{"UUID", "Name", "Confirmed"})
			for _, device := range devices {
				t.AppendRows([]table.Row{{device.ID, device.Name, device.Confirmed}})
			}
			t.Render()
		},
	}
	return cmd
}
