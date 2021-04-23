package util

import (
	"fmt"
	"os"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

// TODO think of a better solution for error management here
func Wrap(cmd func() error) func(*cobra.Command, []string) {
	return func(*cobra.Command, []string) {
		if err := cmd(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}

func ReadPassword(prompt string) string {
	fmt.Print(prompt)
	b, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		panic(err)
	}
	fmt.Println()
	return string(b)
}
