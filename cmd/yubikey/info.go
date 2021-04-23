package yubikey

import (
	"github.com/spf13/cobra"

	"github.com/feeltheajf/ztunnel/cmd/util"
)

var infoCmd = &cobra.Command{
	Use:     "info",
	Aliases: []string{"i"},
	Short:   "Show status of the YubiKey PIV application",
	Run:     util.Wrap(info),
}

func info() error {
	yk, err := open()
	if err != nil {
		return err
	}
	defer yk.Close()
	return printInfo(yk)
}
