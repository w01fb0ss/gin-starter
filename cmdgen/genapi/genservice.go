package genapi

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/w01fb0ss/gin-starter/pkg/gzutil"
)

func (self *generator) MatchRoutesService(content string) (err error) {
	var serviceMap = make(map[string]struct{})
	var serviceRouterMap = make(map[string]struct{})
	var serviceFuncMap = make(map[string]struct{})
	self.services = []*serviceSpec{}
	// serviceRegex := regexp.MustCompile(`service\s+(\w+)\s*{([^}]*)}`)
	// serviceRegex := regexp.MustCompile(`service\s+(\w+)(?:\s+Group\s+([\w,]+))?\s*{([^}]*)}`)
	serviceRegex := regexp.MustCompile(`(?m)(?:\s*@Summary\s+([^\n\r]+))?\s*service\s+(\w+)(?:\s+Group\s+([\w,]+))?\s*{([^}]*)}`)
	serviceMatches := serviceRegex.FindAllStringSubmatch(content, -1)
	for _, serviceMatch := range serviceMatches {
		var service serviceSpec
		service.Summary = serviceMatch[1]
		service.Name = serviceMatch[2]
		service.Group = serviceMatch[3]
		routesBlock := serviceMatch[4]
		if service.Summary == "" {
			service.Summary = service.Name
		}

		routeLineRegex := regexp.MustCompile(`(\w+)\s+([\w/:]+)(?::(\w+))?\s*(?:\(([\[\]\w]+)\))?\s*returns\s*(?:\(([\[\]\w]+)\))?`)
		summaryRegex := regexp.MustCompile(`@Summary\s+(.+)`)
		lines := strings.Split(routesBlock, "\n")
		var lastSummary string

		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			if strings.HasPrefix(line, "@Summary") {
				if matches := summaryRegex.FindStringSubmatch(line); len(matches) == 2 {
					lastSummary = matches[1]
				}
				continue
			}

			if routeMatch := routeLineRegex.FindStringSubmatch(line); len(routeMatch) > 0 {
				uriVal, rustFulVal := gzutil.ConvertRestfulURLToUri(routeMatch[2])
				nameArr := strings.Split(uriVal, "/")
				for i, s := range nameArr {
					nameArr[i] = gzutil.UcFirst(s)
				}
				nameVal := strings.Join(nameArr, "")
				if lastSummary == "" {
					lastSummary = gzutil.UcFirst(service.Name) + nameVal
				}

				nowRouter := &routeSpec{
					Method:       routeMatch[1],
					Path:         routeMatch[2],
					Name:         nameVal,
					RustFulKey:   rustFulVal,
					RequestType:  routeMatch[4],
					ResponseType: routeMatch[5],
					Summary:      lastSummary,
				}

				routerUniqueKey := strings.ToLower(service.Name + nowRouter.Path + nowRouter.Method)
				if _, ok := serviceRouterMap[routerUniqueKey]; ok {
					return fmt.Errorf("router %s is duplicated", nowRouter.Path)
				}

				funcUniqueKey := strings.ToLower(service.Name + nowRouter.Name)
				if _, ok := serviceFuncMap[funcUniqueKey]; ok {
					return fmt.Errorf("func %s is duplicated", nowRouter.Name)
				}
				serviceRouterMap[routerUniqueKey] = struct{}{}
				serviceFuncMap[funcUniqueKey] = struct{}{}
				service.Routes = append(service.Routes, nowRouter)
				lastSummary = ""
			}
		}

		uniqueKey := strings.ToLower(service.Name + service.Group)
		if _, ok := serviceMap[uniqueKey]; ok {
			return fmt.Errorf("Service name_group %s_%s is repeated", service.Name, service.Group)
		}
		serviceMap[uniqueKey] = struct{}{}
		self.services = append(self.services, &service)
	}

	return nil
}
