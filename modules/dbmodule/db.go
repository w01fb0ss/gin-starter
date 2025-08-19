package dbmodule

import (
	"encoding/json"
	"fmt"

	"github.com/soryetong/gooze-starter/gooze"
	"github.com/soryetong/gooze-starter/gzconsole"
	"github.com/soryetong/gooze-starter/pkg/gzutil"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	DbTypeMysql      = "mysql"
	DbTypePostgresql = "postgresql"
	DbTypeSqlite     = "sqlite"
	DbTypeSqlserver  = "sqlserver"
	DbTypeOracle     = "oracle"
)

func init() {
	gzconsole.Register(10, dbCmd)
}

var dbCmd = &cobra.Command{
	Use:    "db",
	Short:  "Init DB",
	Long:   `加载DB模块`,
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return initFunc()
	},
}

type dbConfig struct {
	Name            string
	Dsn             string
	Driver          string
	UseGorm         bool
	LogLevel        int
	EnableLogWriter bool
	MaxIdleConn     int
	MaxConn         int
	SlowThreshold   int
}

func initFunc() error {
	conf := viper.Get(`databases`)
	confMap, ok := conf.([]interface{})
	if !ok || len(confMap) == 0 {
		return fmt.Errorf("请确保 `databases` 模块的配置符合要求")
	}

	isDefault := len(confMap) == 1
	for _, v := range confMap {
		dbConfMap, ok := v.(map[string]interface{})
		if !ok {
			return fmt.Errorf("请确保 `databases` 模块的配置符合要求")
		}

		jsonData, err := json.Marshal(dbConfMap)
		if err != nil {
			return fmt.Errorf("请确保 `databases` 模块的配置符合要求")
		}

		var dbConf dbConfig
		if err = json.Unmarshal(jsonData, &dbConf); err != nil {
			return fmt.Errorf("请确保 `databases` 模块的配置符合要求")
		}

		// 默认值设置
		dbConf.LogLevel = gzutil.Ternary(dbConf.LogLevel <= 0, 3, dbConf.LogLevel)
		dbConf.MaxConn = gzutil.Ternary(dbConf.MaxConn <= 0, 200, dbConf.MaxConn)
		dbConf.MaxIdleConn = gzutil.Ternary(dbConf.MaxIdleConn <= 0, 10, dbConf.MaxIdleConn)
		dbConf.SlowThreshold = gzutil.Ternary(dbConf.SlowThreshold <= 0, 2000, dbConf.SlowThreshold)

		if dbConf.Dsn == "" || dbConf.Name == "" {
			return fmt.Errorf("你正在加载数据库 [%s] 模块，但配置缺少，请先添加配置", dbConf.Name)
		}

		var funcName string
		if dbConf.UseGorm {
			gdb, err := newGormDB(&dbConf)
			if err != nil {
				return err
			}
			gooze.SetDb(dbConf.Name, gdb, nil)
			if isDefault {
				gooze.SetDb("default", gdb, nil)
			}
			funcName = gzutil.Ternary(isDefault, "gooze.Gorm()", fmt.Sprintf(`gooze.Gorm("%s")`, dbConf.Name))
		} else {
			sdb, err := newSqlxDB(&dbConf)
			if err != nil {
				return err
			}
			gooze.SetDb(dbConf.Name, nil, sdb)
			if isDefault {
				gooze.SetDb("default", nil, sdb)
			}
			funcName = gzutil.Ternary(isDefault, "gooze.Sqlx()", fmt.Sprintf(`gooze.Sqlx("%s")`, dbConf.Name))
		}

		gzconsole.Echo.Infof("✅  提示: [%s] DB 模块加载成功, 你可以使用 `%s` 进行数据操作\n", dbConf.Name, funcName)
	}

	return nil
}
