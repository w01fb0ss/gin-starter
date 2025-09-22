package gzconsole

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/common-nighthawk/go-figure"
	"github.com/fatih/color"
)

const bannerBody = `
{{.Desc}}

📘 {{.LabelDocs}}     {{.Docs}}
💻 {{.LabelAuthor}}   {{.Author}}
🔧 {{.LabelVersion}}  {{.Version}}

🌍 {{.Global}} :
{{range .GlobalCommands}}  {{.Name | printf " %-20s"}}  {{.Desc}}
{{end}}
🚀 {{.LabelCommands}} :
{{range .Commands}}  {{.Name | printf "- %-15s"}}  {{.Desc}}
{{end}}
`

type CommandInfo struct {
	Name string
	Desc string
}

type bannerData struct {
	Desc           string
	Docs           string
	Author         string
	Version        string
	LabelDocs      string
	LabelAuthor    string
	LabelVersion   string
	LabelCommands  string
	Commands       []CommandInfo
	Global         string
	GlobalCommands []CommandInfo
}

func Show(commands, globalCommands []CommandInfo) {
	fmt.Println("\n")
	title := figure.NewFigure("Gin Starter", "", true)
	color.New(color.FgHiGreen).Println(title.String())

	data := bannerData{
		Desc:           color.New(color.FgWhite).Sprint("Elegant Go + Gin Scaffold，为了快速构建Go项目而生的应用脚手架"),
		Docs:           "https://github.com/w01fb0ss/gin-starter",
		Author:         "soryetong@gmail.com",
		Version:        "v0.0.1",
		LabelDocs:      "Docs",
		LabelAuthor:    "Author",
		LabelVersion:   "Version",
		LabelCommands:  "Commands",
		Commands:       commands,
		Global:         "Global Flags",
		GlobalCommands: globalCommands,
	}

	tmpl := template.Must(template.New("banner").Parse(bannerBody))
	var out bytes.Buffer
	if err := tmpl.Execute(&out, data); err != nil {
		return
	}

	color.New(color.FgHiWhite).Println(out.String())
}
