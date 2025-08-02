package genmodel

import (
	"github.com/soryetong/gooze-starter/gzconsole"
	"github.com/spf13/cobra"
)

var CmdGen = &cobra.Command{
	Use:   "model",
	Short: "Generate GORM model",
	RunE: func(cmd *cobra.Command, args []string) error {
		gzconsole.Echo.Debugf("即将上线......")

		return nil
	},
}
