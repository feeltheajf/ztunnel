package yubikey

import (
	"errors"

	"github.com/spf13/cobra"

	"github.com/feeltheajf/ztunnel/cmd/util"
)

// TODO reset, set PIN/PUK/MGM
var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Show status of the YubiKey PIV application",
	Run:   util.Wrap(reset),
}

func reset() error {
	return errors.New("not implemented")
}
