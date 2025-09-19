package gzmiddleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/w01fb0ss/gin-starter/base"
	"github.com/w01fb0ss/gin-starter/pkg/gzauth"
	"github.com/w01fb0ss/gin-starter/pkg/gzerror"
)

func Jwt() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tokenString := ctx.GetHeader("Authorization")
		if tokenString == "" || !strings.HasPrefix(tokenString, "Bearer ") {
			base.Fail(ctx, gzerror.NeedLogin)
			ctx.Abort()
			return
		}

		claims, err := gzauth.ParseJwtToken(tokenString[7:])
		if err != nil {
			base.Fail(ctx, gzerror.NeedLogin)
			ctx.Abort()
			return
		}

		ctx.Set("claims", claims)
		ctx.Next()
	}
}
