package cmd

import (
	"github.com/spf13/cobra"

	"github.com/feeltheajf/ztunnel/cmd/client"
	"github.com/feeltheajf/ztunnel/cmd/server"
	"github.com/feeltheajf/ztunnel/cmd/tpm"
	"github.com/feeltheajf/ztunnel/cmd/yubikey"
	"github.com/feeltheajf/ztunnel/config"
)

var cmd = &cobra.Command{
	Use: config.App,
}

func init() {
	cobra.EnableCommandSorting = false

	cmd.AddCommand(client.Cmd)
	cmd.AddCommand(server.Cmd)
	cmd.AddCommand(tpm.Cmd)
	cmd.AddCommand(yubikey.Cmd)
}

func Execute() error {
	return cmd.Execute()
}
