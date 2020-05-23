package main

import (
	"fmt"
	"os"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"
)

// GetPassword gets the password either form an environment variable or interactively
func GetPassword(prompt, env string, interactive bool) string {
	password := os.Getenv(env)

	if interactive && password == "" && terminal.IsTerminal(syscall.Stdin) {
		fmt.Fprintf(os.Stdin, "%s: ", prompt)

		pw, err := terminal.ReadPassword(syscall.Stdin)
		defer os.Stdin.Write([]byte("\n"))

		if err != nil {
			panic(err)
		}

		password = string(pw)
	}

	return password
}
