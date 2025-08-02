package genapi

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/soryetong/gooze-starter/gzconsole"
	"github.com/soryetong/gooze-starter/pkg/gzutil"
)

const serverContentTemplate = `
package {{.PackageName}}

import (
	"github.com/soryetong/gooze-starter/gooze"
	"github.com/soryetong/gooze-starter/pkg/gzutil"
	"github.com/soryetong/gooze-starter/modules/httpmodule"
	"{{ .RouterPackagePath}}"
)

func init() {
	gooze.RegisterService(&{{ .ServerName}}{})
}

type {{ .ServerName}} struct {
	*gooze.IServer

	httpModule httpmodule.IHttp
}

func (self *{{ .ServerName}}) OnStart() (err error) {
	// 添加回调函数
	self.httpModule.OnStop(self.exitCallback())

	{{if .HasViper}} self.httpModule.Init(self, viper.GetString("App.Addr"), 5, router.InitRouter()) {{ else }}
	self.httpModule.Init(self, {{ .ServerAddr}}, {{ .Timeout}}, router.InitRouter()) {{end}}
	err = self.httpModule.Start()

	return
}

// TODO 添加回调函数, 无逻辑可直接删除这个方法
func (self *{{ .ServerName}}) exitCallback() *gzutil.OrderlyMap {
	callback := gzutil.NewOrderlyMap()
	callback.Append("exit", func() {
		gooze.Log.Info("这是程序退出后的回调函数, 执行你想要执行的逻辑, 无逻辑可以直接删除这段代码")
	})
	
	return callback
}
`

func (self *generator) GenServer() (err error) {
	outputName := gzutil.Ternary(self.moduleName == "", "http", self.moduleName)
	serverName := fmt.Sprintf("%sServer", gzutil.UcFirst(outputName))
	path := filepath.Join(self.output, self.moduleName, self.serverPackageName)
	// path := self.serverPackagePath
	if err = os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}
	filename := filepath.Join(path, fmt.Sprintf("%s.go", serverName))
	if _, err = os.Stat(filename); err == nil {
		gzconsole.Echo.Info(fmt.Sprintf("服务文件: %s 已存在，不进行重写", filename))
		return nil
	}

	contentTmpl, err := template.New("server").Parse(serverContentTemplate)
	if err != nil {
		return err
	}

	var builder strings.Builder
	data := map[string]interface{}{
		"PackageName":       self.serverPackageName,
		"ServerName":        serverName,
		"ServerAddr":        "gooze.Config.App.Addr",
		"HasViper":          false,
		"Timeout":           "gooze.Config.App.Timeout",
		"RouterPackagePath": filepath.Join(self.packageName, self.routerPackagePath),
	}
	if err = contentTmpl.Execute(&builder, data); err != nil {
		return err
	}
	builder.WriteString("")

	file, err := os.Create(filename)
	defer file.Close()
	if err != nil {
		return err
	}

	gzconsole.Echo.Info("正在生成服务文件: ", filename)
	if _, err = file.WriteString(builder.String()); err != nil {
		return err
	}
	self.formatFileWithGofmt(filename)

	return nil
}
