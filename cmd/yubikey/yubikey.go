package yubikey

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:     "yubikey",
	Aliases: []string{"y"},
	Short:   "YubiKey management",
}

func init() {
	Cmd.AddCommand(attestCmd)
	Cmd.AddCommand(infoCmd)
	Cmd.AddCommand(resetCmd)
}
