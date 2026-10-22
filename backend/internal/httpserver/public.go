package httpserver

import (
	"fmt"
	"strings"

	"github.com/JRamonCarralero/nexo/backend/internal/httpserver/api"
	"github.com/JRamonCarralero/nexo/backend/internal/httpserver/middleware"
)

// publicRoutes reads the embedded contract and returns the operations that
// declare `security: []`. Everything else requires authentication.
func publicRoutes() (middleware.IsPublic, error) {
	spec, err := api.GetSpec()
	if err != nil {
		return nil, fmt.Errorf("loading embedded OpenAPI spec: %w", err)
	}

	public := map[string]bool{}
	for path, item := range spec.Paths.Map() {
		for method, op := range item.Operations() {
			if op.Security != nil && len(*op.Security) == 0 {
				public[routeKey(method, BasePath+ginPath(path))] = true
			}
		}
	}
	return func(method, route string) bool { return public[routeKey(method, route)] }, nil
}

func routeKey(method, route string) string { return strings.ToUpper(method) + " " + route }

// ginPath converts OpenAPI path parameters ({id}) to gin's syntax (:id).
func ginPath(path string) string {
	var b strings.Builder
	for _, seg := range strings.Split(path, "/") {
		if seg == "" {
			continue
		}
		b.WriteByte('/')
		if strings.HasPrefix(seg, "{") && strings.HasSuffix(seg, "}") {
			b.WriteString(":" + strings.TrimSuffix(strings.TrimPrefix(seg, "{"), "}"))
		} else {
			b.WriteString(seg)
		}
	}
	return b.String()
}
