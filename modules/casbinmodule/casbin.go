package casbinmodule

import (
	"fmt"

	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/w01fb0ss/gin-starter/base"
	"github.com/w01fb0ss/gin-starter/gzconsole"
	"github.com/w01fb0ss/gin-starter/pkg/gzutil"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func init() {
	gzconsole.Register(5, casbinCmd)
}

var casbinCmd = &cobra.Command{
	Use:    "casbin",
	Short:  "Init Casbin",
	Long:   `加载Casbin模块之后，可以通过 base.Casbin 进行权限校验`,
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return initCasbin()
	},
}

func initCasbin() error {
	modePath := viper.GetString("Casbin.ModePath")
	if modePath == "" {
		return fmt.Errorf("Casbin.ModePath 为空，请检查配置文件")
	}

	dbName := viper.GetString("Casbin.DbName")
	dbName = gzutil.Ternary(dbName == "", "default", dbName)
	db := base.Gorm(dbName)
	if db == nil {
		if base.Sqlx(dbName) == nil {
			return fmt.Errorf("casbin 基于 Gorm 实现，请先加载至少一个 databases 模块")
		}
		// 复用 sqlx 的连接池创建一个 GORM 实例
		open, err := gorm.Open(mysql.New(mysql.Config{
			Conn: base.Sqlx(dbName),
		}), &gorm.Config{
			SkipDefaultTransaction: true,
		})
		if err != nil {
			return fmt.Errorf("复用 sqlx 的连接池创建一个 GORM 实例错误: %v", err)
		}
		db = open
	}
	a, _ := gormadapter.NewAdapterByDB(db)
	syncedEnforcer, err := casbin.NewSyncedEnforcer(modePath, a)
	if err != nil {
		return fmt.Errorf("casbin加载失败: %v", err)
	}
	_ = syncedEnforcer.LoadPolicy()

	base.Casbin = syncedEnforcer
	gzconsole.Echo.Info("✅  提示: [Casbin] 模块加载成功, 你可以使用 `base.Casbin` 进行权限操作\n")
	return nil
}
