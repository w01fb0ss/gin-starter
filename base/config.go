package base

import (
	"fmt"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type BaseConfig struct {
	App    app             `mapstructure:"app"`
	Db     []databasesConf `mapstructure:"databases"`
	Redis  redisConf       `mapstructure:"redis"`
	Mongo  mongoConf       `mapstructure:"mongo"`
	Logger logger          `mapstructure:"logger"`
	Casbin casbin          `mapstructure:"casbin"`
	Jwt    jwt             `mapstructure:"jwt"`
	Oss    oss             `mapstructure:"oss"`
}
type app struct {
	Name         string `mapstructure:"name"`
	Env          string `mapstructure:"env"`
	Addr         string `mapstructure:"addr"`
	Timeout      int    `mapstructure:"timeout"`
	RouterPrefix string `mapstructure:"routerPrefix"`
	CacheCap     int    `mapstructure:"cacheCap"`
	CacheShard   int    `mapstructure:"cacheShard"`
	CacheClear   int    `mapstructure:"cacheClear"`
}
type databasesConf struct {
	Name            string `mapstructure:"name"`
	Driver          string `mapstructure:"driver"`
	Dsn             string `mapstructure:"dsn"`
	UseGorm         bool   `mapstructure:"useGorm"`
	LogLevel        int    `mapstructure:"logLevel"`
	EnableLogWriter bool   `mapstructure:"enableLogWriter"`
	MaxIdleConn     int    `mapstructure:"maxIdleConn"`
	MaxConn         int    `mapstructure:"maxConn"`
	SlowThreshold   int    `mapstructure:"slowThreshold"`
}
type redisConf struct {
	Addr      string `mapstructure:"addr"`
	Password  string `mapstructure:"password"`
	Db        int    `mapstructure:"db"`
	IsCluster bool   `mapstructure:"isCluster"`
}
type mongoConf struct {
	URL string `mapstructure:"Url"`
}
type logger struct {
	Path       string `mapstructure:"path"`
	Mode       string `mapstructure:"mode"`
	Logrotate  bool   `mapstructure:"logrotate"`
	Recover    bool   `mapstructure:"recover"`
	MaxSize    int    `mapstructure:"maxSize"`
	MaxBackups int    `mapstructure:"maxBackups"`
	MaxAge     int    `mapstructure:"maxAge"`
	Compress   bool   `mapstructure:"compress"`
}
type casbin struct {
	ModePath string `mapstructure:"modePath"`
}
type jwt struct {
	SecretKey string `mapstructure:"secretKey"`
	Expire    int    `mapstructure:"expire"`
}

type oss struct {
	Type       string `mapstructure:"type"`
	SavePath   string `mapstructure:"savePath"`
	Url        string `mapstructure:"url"`
	AccessKey  string `mapstructure:"accessKey"`
	SecretKey  string `mapstructure:"secretKey"`
	BucketName string `mapstructure:"bucketName"`
}

func LoadConfig[T any](file string, env string, target *T) error {
	viper.SetConfigFile(file)
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("读取配置文件错误: %s", err)
	}

	if env != "" {
		if err := godotenv.Load(env); err != nil {
			return fmt.Errorf("读取环境变量错误: %s", err)
		}
	}

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 强制初始化必要的配置
	if viper.GetInt("App.CacheCap") == 0 {
		viper.Set("App.CacheCap", 100000)
	}
	if viper.GetInt("App.CacheShard") == 0 {
		viper.Set("App.CacheShard", 64)
	}
	if viper.GetString("Casbin.DbName") == "" {
		viper.Set("Casbin.DbName", "default")
	}

	return viper.Unmarshal(target)
}
