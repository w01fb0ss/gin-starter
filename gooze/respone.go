package gooze

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/soryetong/gooze-starter/pkg/gzerror"
	"github.com/spf13/cast"

	"net/http"
	"time"
)

type PageResult struct {
	List        interface{} `json:"list"`
	Total       int64       `json:"total"`
	CurrentPage int64       `json:"current_page"`
	PageSize    int64       `json:"page_size"`
}

type Response struct {
	Code    int64       `json:"code"`
	Msg     string      `json:"msg"`
	Data    interface{} `json:"data"`
	NowTime int64       `json:"nowTime"`
	UseTime string      `json:"useTime"`
}

func result(ctx *gin.Context, code int64, data interface{}, msg string) {
	resp := Response{
		Code:    code,
		Msg:     msg,
		Data:    data,
		NowTime: time.Now().Unix(),
	}
	if useTime(ctx) != "" {
		resp.UseTime = useTime(ctx)
	}
	ctx.JSON(http.StatusOK, resp)
}

func Success(ctx *gin.Context, data interface{}) {
	result(ctx, gzerror.OK, data, "success")
}

func SuccessWithMessage(ctx *gin.Context, msg string) {
	result(ctx, gzerror.OK, nil, msg)
}

func Fail(ctx *gin.Context, code int64, msg ...string) {
	var message string
	if len(msg) > 0 {
		message = msg[0]
	} else {
		message = gzerror.GetErrorMessage(code)
	}

	result(ctx, code, nil, message)
}

func FailWithMessage(ctx *gin.Context, msg string) {
	result(ctx, gzerror.Error, nil, msg)
}

func useTime(c *gin.Context) string {
	startTime, _ := c.Get("requestStartTime")
	stopTime := time.Now().UnixMicro()
	if startTime == nil {
		return ""
	}

	return fmt.Sprintf("%.6f", float64(stopTime-cast.ToInt64(startTime))/1000000)
}
