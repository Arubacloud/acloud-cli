package cmd

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

// confirmDelete prompts the user for confirmation before a destructive operation.
// Returns true if the user confirmed, false if they declined or stdin is non-interactive.
func confirmDelete(resourceType, id string) (bool, error) {
	fi, err := os.Stdin.Stat()
	if err != nil || (fi.Mode()&os.ModeCharDevice) == 0 {
		return false, fmt.Errorf("delete requires --yes/-y in non-interactive mode")
	}
	fmt.Printf("Are you sure you want to delete %s %s? (yes/no): ", resourceType, id)
	var response string
	fmt.Scanln(&response)
	if response != "yes" && response != "y" {
		fmt.Println("Delete cancelled")
		return false, nil
	}
	return true, nil
}

// promptConfirmOptional asks a yes/no question when stdin is an interactive
// terminal. Returns false silently in non-interactive mode — the caller treats
// that as "no" and moves on without erroring.
func promptConfirmOptional(question string) bool {
	fi, err := os.Stdin.Stat()
	if err != nil || (fi.Mode()&os.ModeCharDevice) == 0 {
		return false
	}
	fmt.Printf("%s (y/N): ", question)
	var response string
	fmt.Scanln(&response)
	return response == "yes" || response == "y"
}

// readSecret prompts the user for a secret value with echo disabled.
// Returns an error if stdin is not an interactive terminal.
func readSecret(prompt string) (string, error) {
	fi, err := os.Stdin.Stat()
	if err != nil || (fi.Mode()&os.ModeCharDevice) == 0 {
		return "", fmt.Errorf("cannot read secret interactively: stdin is not a terminal; set ACLOUD_CLIENT_SECRET instead")
	}
	fmt.Fprint(os.Stderr, prompt)
	secret, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", fmt.Errorf("reading secret: %w", err)
	}
	return string(secret), nil
}
