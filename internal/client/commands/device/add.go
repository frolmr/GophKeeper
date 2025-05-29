// add implements the device add command.
//
// This handles:
// - New device registration
package device

import (
	"fmt"

	"github.com/frolmr/GophKeeper/internal/client/app"
	"github.com/frolmr/GophKeeper/internal/client/app/device"
	"github.com/spf13/cobra"
)

// NewAddCommand creates the 'device' add command.
//
// Returns:
//
//	*cobra.Command: The configured add subcommand
func NewAddCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Register a new users device",
		Long:  "Interactive command to register a new users device",
		Run: func(cmd *cobra.Command, args []string) {
			needAuth := true
			needCrypto := true

			app, err := app.NewApp(needAuth, needCrypto)
			if err != nil {
				fmt.Printf("Failed to setup cli: %v", err)
			}
			defer app.Close()

			service := device.NewDeviceService(app.Storage, app.CryptoService, app.ConnAdapter)

			if err := service.AddDevice(); err != nil {
				fmt.Printf("Add device command failed: %v", err.Error())
			} else {
				fmt.Println("Device added successful!")
			}
		},
	}
	return cmd
}
