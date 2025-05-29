// login implements the user login command.
//
// This handles:
// - Login inputs
package user

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/frolmr/GophKeeper/internal/client/app"
	"github.com/frolmr/GophKeeper/internal/client/app/user"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// NewLoginCommand creates the 'user' login command.
//
// Returns:
//
//	*cobra.Command: The configured login subcommand
func NewLoginCommand() *cobra.Command {
	var email, password string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login user",
		Long:  "Interactive command to login the user. This will prompt you for email and password.",
		Run: func(cmd *cobra.Command, args []string) {
			if email == "" || password == "" {
				reader := bufio.NewReader(os.Stdin)

				fmt.Print("Enter Email: ")
				email, _ = reader.ReadString('\n')
				email = strings.TrimSpace(email)

				fmt.Print("Enter Password: ")
				bytePassword1, _ := term.ReadPassword(syscall.Stdin)
				fmt.Println()

				password = strings.TrimSpace(string(bytePassword1))
			}

			needAuth := false
			needCrypto := false

			app, err := app.NewApp(needAuth, needCrypto)
			if err != nil {
				fmt.Printf("Failed to setup cli: %v", err)
			}
			defer app.Close()

			service := user.NewUserService(app.Storage, app.CryptoService, app.ConnAdapter)

			if err := service.Login(email, password); err != nil {
				fmt.Printf("User login failed: %v", err.Error())
			} else {
				fmt.Println("Login successful!")
			}
		},
	}
	cmd.Flags().StringVarP(&email, "email", "e", "", "Email address")
	cmd.Flags().StringVarP(&password, "password", "p", "", "Password")

	return cmd
}
