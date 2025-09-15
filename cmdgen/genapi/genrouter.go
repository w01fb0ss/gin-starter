package genapi

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/w01fb0ss/gin-starter/cmdgen/config"
	"github.com/w01fb0ss/gin-starter/gzconsole"
	"github.com/w01fb0ss/gin-starter/pkg/gzutil"
)

const routerContentTemplate = `
package router

import (
	"github.com/gin-gonic/gin"
	"{{ .HandlerPackPath}}"
)

func Init{{.NowGroupName}}Router(routerGroup *gin.RouterGroup) {
{{range .Routes}}{{.GroupName}}Group := routerGroup.Group("/{{.RouteGroup}}")
{{"{"}}{{range .Routes}}
	{{.GroupName}}Group.{{.Method}}("/{{.Path}}", {{.HandlerPackName}}.{{.HandlerName}}){{end}}
{{"}"}}{{end}}
}
`

const routerFuncContent = `
func Init{{.NowGroupName}}Router(routerGroup *gin.RouterGroup) {
{{range .Routes}}{{.GroupName}}Group := routerGroup.Group("/{{.RouteGroup}}")
{{"{"}}{{range .Routes}}
	{{.GroupName}}Group.{{.Method}}("/{{.Path}}", {{.HandlerPackName}}.{{.HandlerName}}){{end}}
{{"}"}}{{end}}
}
`

type RouteTemplateData struct {
	NowGroupName    string
	HandlerPackPath string
	Routes          []RouteGroupTemplateData
}

type RouteGroupTemplateData struct {
	GroupName  string
	RouteGroup string
	Middleware []RouteMiddleware
	Routes     []RouteSpecTemplateData
}

type RouteSpecTemplateData struct {
	GroupName       string
	Method          string
	Path            string
	HandlerPackName string
	HandlerName     string
}

type RouteMiddleware struct {
	GroupName string
	InGina    bool
	Name      string
}

func extractInitFunctions(filename string) ([]string, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	re := regexp.MustCompile(`func (Init\w+)\s*\(`)
	matches := re.FindAllStringSubmatch(string(content), -1)

	var functions []string
	for _, match := range matches {
		if len(match) > 1 {
			functions = append(functions, match[1])
		}
	}
	return functions, nil
}

func cleanRouterBlock(filename string, invalidFunc string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	lines := strings.Split(string(data), "\n")

	var newLines []string
	var insideBlock bool
	var blockStartIdx, blockEndIdx int
	var buffer []string

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		if !insideBlock && strings.Contains(trimmed, ":= r.Group(") {
			blockStartIdx = i
			buffer = []string{line}

			if i+1 < len(lines) && strings.Contains(strings.TrimSpace(lines[i+1]), ".Use(") {
				buffer = append(buffer, lines[i+1])
				i++

				if i+1 < len(lines) && strings.TrimSpace(lines[i+1]) == "{" {
					buffer = append(buffer, lines[i+1])
					i++
					insideBlock = true
					continue
				}
			}
		}

		if insideBlock {
			if trimmed == "}" {
				blockEndIdx = i
				bodyLines := buffer[3:]
				bodyLines = append(bodyLines, lines[blockStartIdx+len(buffer):blockEndIdx]...)

				filteredBody := []string{}
				for _, l := range bodyLines {
					if !strings.Contains(l, invalidFunc+"(") {
						filteredBody = append(filteredBody, l)
					}
				}

				if len(filteredBody) == 0 {
				} else {
					newLines = append(newLines, buffer[:3]...)
					newLines = append(newLines, filteredBody...)
					newLines = append(newLines, lines[i])
				}
				insideBlock = false
				buffer = nil
			}
			continue
		}

		if !insideBlock {
			newLines = append(newLines, line)
		}
	}

	return os.WriteFile(filename, []byte(strings.Join(newLines, "\n")), 0644)
}

func (self *generator) GenRouter() (err error) {
	filename := filepath.Join(self.output, self.moduleName, self.routerPackageName, self.fileName)
	enterGoFileName := filepath.Join(self.output, self.moduleName, self.routerPackageName, "router.go")
	// filename := filepath.Join(self.routerPackagePath, self.fileName)
	nowFuncs := []string{}
	for _, service := range self.services {
		templateData := RouteTemplateData{
			HandlerPackPath: filepath.Join(self.packageName, self.handlerPackagePath),
			Routes:          []RouteGroupTemplateData{},
		}
		templateData.NowGroupName = gzutil.UcFirst(service.Name) + gzutil.UcFirst(service.Group)
		newGroupName := strings.ToLower(service.Name[:1]) + service.Name[1:]
		// split := strings.Split(ginahelper.SeparateCamel(service.Name, "/"), "/")
		// newFileName = strings.ToLower(strings.Join(split, "_"))
		group := RouteGroupTemplateData{
			GroupName:  newGroupName,
			RouteGroup: gzutil.SeparateCamel(service.Name, "/"),
			Routes:     []RouteSpecTemplateData{},
		}

		for _, route := range service.Routes {
			routeData := RouteSpecTemplateData{
				GroupName:       newGroupName,
				Method:          strings.ToUpper(route.Method),
				Path:            strings.ToLower(route.Path),
				HandlerName:     gzutil.UcFirst(service.Name) + route.Name,
				HandlerPackName: self.handlerPackageName,
			}
			group.Routes = append(group.Routes, routeData)
		}
		templateData.Routes = append(templateData.Routes, group)

		// filename := filepath.Join(nowRouterPath, "router.go")
		_, err = os.Stat(filename)
		if os.IsNotExist(err) {
			if err = os.MkdirAll(filepath.Dir(filename), os.ModePerm); err != nil {
				return err
			}
			var builder strings.Builder
			tmpl, err := template.New("router").Parse(routerContentTemplate)
			if err != nil {
				return err
			}

			if err = tmpl.Execute(&builder, templateData); err != nil {
				return err
			}

			// filename := filepath.Join(nowRouterPath, fmt.Sprintf("%s.go", newFileName))
			file, err := os.Create(filename)
			defer file.Close()
			if err != nil {
				return err
			}

			gzconsole.Echo.Info("正在生成路由文件: ", filename)
			if _, err = file.WriteString(builder.String()); err != nil {
				return err
			}
		} else {
			// 文件已存在时，需要替换现有Init函数
			// 1. 读出现有文件内容
			bytes, err := os.ReadFile(filename)
			if err != nil {
				return err
			}
			content := string(bytes)

			// 2. 渲染出新的Init函数
			var builder strings.Builder
			tmpl, err := template.New("routerFunc").Parse(routerFuncContent)
			if err != nil {
				return err
			}

			if err = tmpl.Execute(&builder, templateData); err != nil {
				return err
			}

			newInitFunc := builder.String()

			// 3. 检查是不是已有Init函数
			re := regexp.MustCompile(
				`func Init` + templateData.NowGroupName + `Router\s*\([^)]*\)\s*\{[\s\S]*?\n\}`,
			)
			if re.FindStringIndex(content) == nil {
				content += "\n" + newInitFunc + "\n"
			} else {
				content = re.ReplaceAllString(content, newInitFunc)
			}

			// 4. 写回文件
			if err = os.WriteFile(filename, []byte(content), 0644); err != nil {
				return err
			}

			gzconsole.Echo.Info("正在向现有路由文件中添加Init函数到末尾: ", filename)
		}
		self.formatFileWithGofmt(filename)

		// 更新入口文件
		funcName := fmt.Sprintf("Init%sRouter", templateData.NowGroupName)
		err = self.updateEnterGo(enterGoFileName, funcName, service.Group)
		if err != nil {
			return err
		}

		nowFuncs = append(nowFuncs, funcName)
	}

	functions, err := extractInitFunctions(filename)
	if err != nil {
		return err
	}

	for _, funcName := range functions {
		if !gzutil.InArray(funcName, nowFuncs) {
			gzconsole.Echo.Info("✅ 移除无效的Router函数: ", funcName)
			if err = removeRouterFuncFromFile(filename, funcName); err != nil {
				return err
			}
			if err = cleanRouterBlock(enterGoFileName, funcName); err != nil {
				return err
			}

		}
	}

	self.formatFileWithGofmt(filename)
	self.formatFileWithGofmt(enterGoFileName)

	return nil
}

const enterGoTemplate = `package router

import (
	"github.com/gin-gonic/gin"
	"github.com/w01fb0ss/gin-starter/pkg/gzmiddleware"
	"github.com/spf13/viper"
	"net/http"
)

func InitRouter() *gin.Engine {
	setMode()

	r := gin.Default()
	fs := "/static"
	r.StaticFS(fs, http.Dir("./"+fs))

	r.Use(gzmiddleware.Begin()).Use(gzmiddleware.Cross()){{if .NeedRequestLog}}.Use(gzmiddleware.RequestLog()){{end}}
	publicGroup := r.Group("{{ .RouterPrefix}}")
	{
		// 健康监测
		publicGroup.GET("/health", func(c *gin.Context) {
			c.JSON(200, "ok")
		})

		{{ if eq .GroupName "Public" }}{{.InitPublicFunctions}}(publicGroup){{ end }}
	}

	{{ if eq .GroupName "Auth" }}
	privateAuthGroup := r.Group("{{ .RouterPrefix}}")
	privateAuthGroup.Use(gzmiddleware.Jwt()).Use(gzmiddleware.Casbin())
	{
		{{.InitPrivateAuthFunctions}}(privateAuthGroup)
	}{{ end }}

	{{ if eq .GroupName "Token" }}
	privateTokenGroup := r.Group("{{ .RouterPrefix}}")
	privateTokenGroup.Use(gzmiddleware.Jwt())
	{
		{{.InitPrivateTokenFunctions}}(privateTokenGroup)
	}{{ end }}

	return r
}

func setMode() {
	switch viper.GetString("App.Env") {
	case gin.DebugMode:
		gin.SetMode(gin.DebugMode)
	case gin.ReleaseMode:
		gin.SetMode(gin.ReleaseMode)
	default:
		gin.SetMode(gin.TestMode)
	}
}
`

type EnterGoTemplateData struct {
	RouterPrefix              string
	NeedRequestLog            bool
	GroupName                 string
	InitPublicFunctions       string
	InitPrivateAuthFunctions  string
	InitPrivateTokenFunctions string
}

func (self *generator) updateEnterGo(enterGoFileName, newRouter, nowGroup string) (err error) {
	//var nowGroup string
	//for _, service := range self.services {
	//	nowGroup = service.Group
	//}

	filename := enterGoFileName
	_, err = os.Stat(filename)

	templateData := EnterGoTemplateData{
		GroupName:      nowGroup,
		RouterPrefix:   self.routerPrefix,
		NeedRequestLog: self.needRequestLog,
	}
	if os.IsNotExist(err) {
		if err = os.MkdirAll(filepath.Dir(filename), os.ModePerm); err != nil {
			return err
		}
		file, err := os.Create(filename)
		if err != nil {
			return err
		}

		defer file.Close()
		switch nowGroup {
		case config.Group_Public:
			templateData.InitPublicFunctions = newRouter
		case config.Group_Auth:
			templateData.InitPrivateAuthFunctions = newRouter
		case config.Group_Token:
			templateData.InitPrivateTokenFunctions = newRouter
		default:
		}

		var builder strings.Builder
		tmpl, err := template.New("routerEnter").Parse(enterGoTemplate)
		if err != nil {
			return err
		}
		if err = tmpl.Execute(&builder, templateData); err != nil {
			return err
		}

		if _, err = file.WriteString(builder.String()); err != nil {
			return err
		}

		gzconsole.Echo.Info("✅ 已完成路由入口文件更新: ", filename, "  ", nowGroup)
		// self.formatFileWithGofmt(filename)

		return err
	}

	content, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	groupName := strings.ToLower(nowGroup)
	if nowGroup != config.Group_Public {
		groupName = "private" + nowGroup
	}

	lines := strings.Split(string(content), "\n")
	var newContent []string
	foundGroup := false
	inserted := false
	functionExists := false
	groupStartIndex := -1
	groupEndIndex := -1
	returnIndex := -1

	for i, line := range lines {
		newContent = append(newContent, line)
		if strings.TrimSpace(line) == "return r" {
			returnIndex = i
		}
		if strings.Contains(line, newRouter+"(") {
			functionExists = true
		}
		if strings.Contains(line, groupName+"Group := r.Group(") {
			foundGroup = true
			groupStartIndex = i
		}
		if foundGroup && strings.TrimSpace(line) == "}" {
			groupEndIndex = i
			foundGroup = false
		}
	}

	if groupStartIndex == -1 {
		if returnIndex != -1 {
			newContent = append(newContent[:returnIndex], fmt.Sprintf("\t%sGroup := r.Group(\"%s\")", groupName, self.routerPrefix))
			if nowGroup == config.Group_Auth {
				newContent = append(newContent, "\t"+groupName+"Group.Use(gzmiddleware.Jwt()).Use(gzmiddleware.Casbin())")
			} else if nowGroup == config.Group_Token {
				newContent = append(newContent, "\t"+groupName+"Group.Use(gzmiddleware.Jwt())")
			}
			newContent = append(newContent, "\t{")
			newContent = append(newContent, "\t\t"+newRouter+"("+groupName+"Group)")
			newContent = append(newContent, "\t}\n")
			newContent = append(newContent, lines[returnIndex:]...)
		}
		inserted = true
	} else if !functionExists && groupEndIndex != -1 {
		newContent = append(newContent[:groupEndIndex], "\t\t"+newRouter+"("+groupName+"Group)")
		newContent = append(newContent, lines[groupEndIndex:]...)
		inserted = true
	}

	if inserted {
		// 写回文件
		if err = os.WriteFile(filename, []byte(strings.Join(newContent, "\n")), 0644); err != nil {
			return err
		}
	}

	gzconsole.Echo.Info("✅ 已完成路由入口文件更新: ", filename, "  ", nowGroup)
	// self.formatFileWithGofmt(filename)

	return err
}

func removeRouterFuncFromFile(filename string, funcName string) error {
	contentBytes, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	lines := strings.Split(string(contentBytes), "\n")

	var result []string
	inFunc := false
	bracketDepth := 0

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trim := strings.TrimSpace(line)

		if !inFunc && strings.HasPrefix(trim, "func "+funcName+"(") {
			inFunc = true

			if strings.Contains(line, "{") {
				bracketDepth += strings.Count(line, "{")
				bracketDepth -= strings.Count(line, "}")
			}

			continue
		}

		if inFunc {
			bracketDepth += strings.Count(line, "{")
			bracketDepth -= strings.Count(line, "}")

			if bracketDepth <= 0 {
				inFunc = false
			}
			continue
		}

		result = append(result, line)
	}

	final := removeExtraBlankLines(result)

	return os.WriteFile(filename, []byte(strings.Join(final, "\n")), 0644)
}

func removeExtraBlankLines(lines []string) []string {
	var result []string
	blankCount := 0

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			blankCount++
			if blankCount > 1 {
				continue
			}
		} else {
			blankCount = 0
		}
		result = append(result, line)
	}
	return result
}
