package genapi

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func (self *generator) MatchDto(content string) (err error) {
	self.dtoContents = []*dtoContentsSpec{}
	structRegex := regexp.MustCompile(`type\s+(\w+)\s*{([^}]*)}`)
	matches := structRegex.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		name := match[1]
		fieldsBlock := match[2]
		fields := []*dtoFieldSpec{}

		lines := strings.Split(fieldsBlock, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line) // Remove any leading/trailing whitespace
			if line == "" {
				continue
			}
			field, err := self.parseDtoFieldDeclaration(line)
			if err != nil {
				continue
			}

			fields = append(fields, field)
		}

		self.dtoContents = append(self.dtoContents, &dtoContentsSpec{
			Name:   name,
			Fields: fields,
		})
	}

	return nil
}

func (self *generator) parseDtoFieldDeclaration(declaration string) (*dtoFieldSpec, error) {
	regexPattern := `^\s*(?:(\w+)\s+)?([\w\*\[\]]+)(?:\s+` + "`" + `([^` + "`" + `]*)` + "`" + `)?\s*$`
	re := regexp.MustCompile(regexPattern)

	matches := re.FindStringSubmatch(declaration)
	if matches == nil || len(matches) < 3 {
		return nil, fmt.Errorf("invalid field declaration format: %s", declaration)
	}

	field := &dtoFieldSpec{
		Name: matches[1],
		Type: matches[2],
	}

	// 处理标签（如果存在）
	if len(matches) > 3 && matches[3] != "" {
		field.Tag = matches[3]
	}

	// 处理注释（如果存在）
	if len(matches) > 4 && matches[4] != "" {
		field.Comment = strings.TrimSpace(matches[4])
	}

	return field, nil
}

func (self *generator) GenDto() (err error) {
	filename := filepath.Join(self.output, self.moduleName, self.dtoPackageName, self.fileName)
	if err = os.MkdirAll(filepath.Dir(filename), os.ModePerm); err != nil {
		return err
	}

	for _, dtoContents := range self.dtoContents {
		if err = self.writeToTypeFile(filename, self.formatStruct(dtoContents)); err != nil {
			return err
		}
	}

	return nil
}

func (self *generator) formatStruct(s *dtoContentsSpec) string {
	structDef := fmt.Sprintf("type %s struct {\n", s.Name)
	for _, field := range s.Fields {
		tag := ""
		if field.Tag != "" {
			tag = fmt.Sprintf(" `%s`", field.Tag)
		}

		structDef += fmt.Sprintf("\t%s %s%s\n", field.Name, field.Type, tag)

	}
	structDef += "}\n\n"

	return structDef
}

// writeToTypeFile writes or updates a Go struct in the specified file.
func (self *generator) writeToTypeFile(filename, structContent string) error {
	structNameRegex := regexp.MustCompile(`(?m)^type\s+(\w+)\s+struct\b`)
	matches := structNameRegex.FindStringSubmatch(structContent)
	if len(matches) < 2 {
		return fmt.Errorf("struct name not found in struct definition")
	}
	structName := matches[1]
	existingContent, err := os.ReadFile(filename)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	existingCode := string(existingContent)
	structRegex := regexp.MustCompile(fmt.Sprintf(`(?ms)^type\s+%s\s+struct\s*\{.*?\}\s*(?:\n|$)`, structName))
	cleanedContent := structRegex.ReplaceAllString(existingCode, "")

	var buffer bytes.Buffer
	if !strings.Contains(existingCode, "package ") {
		buffer.WriteString(fmt.Sprintf("package %s\n\n", self.dtoPackageName))
	}

	if strings.TrimSpace(cleanedContent) != "" {
		buffer.WriteString(strings.TrimSpace(cleanedContent))
		buffer.WriteString("\n\n")
	}

	buffer.WriteString(strings.TrimSpace(structContent))
	buffer.WriteString("\n")

	formattedCode, err := format.Source(buffer.Bytes())
	if err != nil {
		return fmt.Errorf("failed to format Go code: %w\nGenerated Code:\n%s", err, buffer.String())
	}

	return os.WriteFile(filename, formattedCode, 0644)
}
