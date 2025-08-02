package cmdgen

import (
	"github.com/soryetong/gooze-starter/cmdgen/genapi"
	"github.com/soryetong/gooze-starter/cmdgen/gencurd"
	"github.com/soryetong/gooze-starter/cmdgen/genmodel"
	"github.com/soryetong/gooze-starter/gzconsole"
	"github.com/spf13/cobra"
)

func init() {
	genCmd.AddCommand(genapi.CmdGen)
	genCmd.AddCommand(genmodel.CmdGen)
	genCmd.AddCommand(gencurd.CmdGen)
	gzconsole.Register(60, genCmd)
}

var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "Code generator entry",
	Long: `Code generation entry point for gooze-starter.
		Available Subcommands:
		  -  api       Generate Gin route & handler based on .api spec file
			  --src       Path to the .api description file (required)
			  --output    Output directory for generated handler files (required)
			  --log       Open request Log (default: false)
		
		  -  model     Generate GORM model from database schema
			  --db-name       Database instance name (from config)
			  --table         Table name to generate model for
		
		  -  crud      Generate full CRUD module (model + handler + routes)
			  --name          Module name
			  --with-api      Whether to include API endpoints (default: true)
		`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}
