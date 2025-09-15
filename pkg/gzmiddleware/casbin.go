package gzmiddleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
	"github.com/w01fb0ss/gin-starter/gooze"
	"github.com/w01fb0ss/gin-starter/gzconsole"
	"github.com/w01fb0ss/gin-starter/pkg/gzauth"
	"github.com/w01fb0ss/gin-starter/pkg/gzerror"
	"github.com/w01fb0ss/gin-starter/pkg/gzutil"
)

func Casbin() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if strings.ToLower(ctx.Request.Method) == "OPTIONS" {
			ctx.Next()
			return
		}

		roleId := gzauth.GetTokenValue[int64](ctx, "role_id")
		if roleId == 0 {
			gzconsole.Echo.Info("ℹ️ 提示: 无法使用 `Casbin` 权限校验, 请确保 `Token` 中包含了字段 `role_id`")
			ctx.Next()
			return
		}

		path := gzutil.ConvertToRestfulURL(strings.TrimPrefix(ctx.Request.URL.Path, viper.GetString("App.RouterPrefix")))
		success, _ := gooze.Casbin.Enforce(cast.ToString(roleId), path, ctx.Request.Method)
		if !success {
			gooze.Fail(ctx, gzerror.NoAuth)
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}
