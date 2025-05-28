// confirm implements the device confirm command.
//
// This handles:
// - New device confirmation
package device

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/frolmr/GophKeeper/internal/client/app"
	"github.com/frolmr/GophKeeper/internal/client/app/device"
	"github.com/spf13/cobra"
)

// NewConfirmCommand creates the 'device' confirm command.
//
// Returns:
//
//	*cobra.Command: The configured confirm subcommand
func NewConfirmCommand() *cobra.Command {
	var deviceID string

	cmd := &cobra.Command{
		Use:   "confirm",
		Short: "Confirm users device",
		Long:  "Interactive command to confirm unconfirmed users device",
		Run: func(cmd *cobra.Command, args []string) {
			if deviceID == "" {
				reader := bufio.NewReader(os.Stdin)
				fmt.Print("Enter Device UUID: ")
				deviceID, _ = reader.ReadString('\n')
				deviceID = strings.TrimSpace(deviceID)
			}

			needAuth := true
			needCrypto := true

			app, err := app.NewApp(needAuth, needCrypto)
			if err != nil {
				fmt.Printf("Failed to setup cli: %v", err)
			}
			defer app.Close()

			service := device.NewDeviceService(app.Storage, app.CryptoService, app.ConnAdapter)

			if err := service.ConfirmDevice(deviceID); err != nil {
				fmt.Printf("Device confirmation failed: %v", err.Error())
			} else {
				fmt.Println("Device confirmed")
			}
		},
	}
	cmd.Flags().StringVarP(&deviceID, "device_id", "i", "", "Device UUID")

	return cmd
}
