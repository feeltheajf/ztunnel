package tpm

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:     "tpm",
	Aliases: []string{"t"},
	Short:   "TPM management",
}

func init() {
	Cmd.AddCommand(attestCmd)
}
