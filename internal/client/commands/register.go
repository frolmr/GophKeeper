package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/frolmr/GophKeeper/pkg/validator"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

//nolint:gochecknoinits // need for command module
func init() {
	rootCmd.AddCommand(registerCmd)
	registerCmd.Flags().StringP("email", "e", "", "Email address")
	registerCmd.Flags().StringP("password", "p", "", "Password")
}

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register a new user account",
	Long: `Interactive command to register a new user.
This will prompt you for email and password.`,
	Run: func(cmd *cobra.Command, args []string) {
		email, _ := cmd.Flags().GetString("email")
		password, _ := cmd.Flags().GetString("password")

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

		registerUser(email, password)
	},
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

func registerUser(email, password string) {
	if err := gk.UserService.Register(email, password); err != nil {
		fmt.Println("User registration failed: ", err.Error())
	}
	fmt.Println("Registration successful!")
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
