package mongomodule

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/w01fb0ss/gin-starter/base"
	"github.com/w01fb0ss/gin-starter/gzconsole"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func init() {
	gzconsole.Register(20, mongoCmd)
}

var mongoCmd = &cobra.Command{
	Use:   "mongoDB",
	Short: "Init MongoDB",
	Long:  `加载MongoDB模块之后，可以通过 base.Mdb 进行数据操作`,
	RunE: func(cmd *cobra.Command, args []string) error {
		url := viper.GetString("Mongo.Url")
		if url == "" {
			return fmt.Errorf("你正在加载MongoDB模块，但是你未配置Mongo.Url，请先添加配置")
		}

		return initClient(url)
	},
}

func initClient(url string) error {
	clientOptions := options.Client().ApplyURI(url)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return fmt.Errorf("MongoDB连接失败: %w", err)
	}

	// 检查连接
	if err = client.Ping(context.TODO(), nil); err != nil {
		return fmt.Errorf("MongoDB连接失败: %w", err)
	}

	base.Mdb = client
	gzconsole.Echo.Info("✅ 提示: [Mongo] 模块加载成功, 你可以使用 `base.Mdb` 进行数据操作\n")
	return nil
}
