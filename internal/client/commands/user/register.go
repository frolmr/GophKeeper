// register implements the user registration command.
//
// This handles:
// - Validating registration inputs
// - Initial device registration
package user

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/frolmr/GophKeeper/internal/client/app"
	"github.com/frolmr/GophKeeper/internal/client/app/user"
	"github.com/frolmr/GophKeeper/pkg/validator"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// NewRegisterCommand creates the 'user' register command.
//
// Returns:
//
//	*cobra.Command: The configured register subcommand
func NewRegisterCommand() *cobra.Command {
	var email, password string

	cmd := &cobra.Command{
		Use:   "register",
		Short: "Register a new user account",
		Long: `Interactive command to register a new user.
This will prompt you for email and password.`,
		Run: func(cmd *cobra.Command, args []string) {
			if email == "" || password == "" {
				email, password = getCredentials()
			} else {
				if !validator.IsValidEmail(email) {
					printInvalidEmailMessage()
					return
				}
				if !validator.IsPasswordStrong(password) {
					printInvalidPasswordMessage()
					return
				}
			}

			needAuth := false
			needCrypto := true

			app, err := app.NewApp(needAuth, needCrypto)
			if err != nil {
				fmt.Printf("Failed to setup cli: %v", err)
			}
			defer app.Close()

			service := user.NewUserService(app.Storage, app.CryptoService, app.ConnAdapter)

			if err := service.Register(email, password); err != nil {
				fmt.Printf("User registration failed: %v", err.Error())
			} else {
				fmt.Println("Registration successful!")
			}
		},
	}
	cmd.Flags().StringVarP(&email, "email", "e", "", "Email address")
	cmd.Flags().StringVarP(&password, "password", "p", "", "Password")

	return cmd
}

func getCredentials() (email, password string) {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("Enter Email: ")
		email, _ := reader.ReadString('\n')
		email = strings.TrimSpace(email)
		if !validator.IsValidEmail(email) {
			fmt.Println("Invalid email format. Please try again.")
			continue
		}
		for {
			fmt.Print("Enter Password: ")
			bytePassword1, _ := term.ReadPassword(syscall.Stdin)
			fmt.Println()

			fmt.Print("Confirm Password: ")
			bytePassword2, _ := term.ReadPassword(syscall.Stdin)
			fmt.Println()

			if string(bytePassword1) != string(bytePassword2) {
				fmt.Println("Passwords don't match. Please try again.")
				continue
			}

			password := strings.TrimSpace(string(bytePassword1))
			if !validator.IsPasswordStrong(password) {
				printInvalidPasswordMessage()
				continue
			}

			return email, password
		}
	}
}

func printInvalidEmailMessage() {
	fmt.Println("Invalid email format. Please try again.")
}

func printInvalidPasswordMessage() {
	fmt.Println("Password is not strong enough. It must contain:")
	fmt.Println("- at least 8 characters")
	fmt.Println("- uppercase letter(s)")
	fmt.Println("- lowercase letter(s)")
	fmt.Println("- number(s)")
	fmt.Println("- special character(s)")
}
