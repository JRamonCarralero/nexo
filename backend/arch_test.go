package backend_test

// Architecture rules (see D-008). This test reads the real imports of every
// package and fails when a dependency crosses a forbidden boundary.

import (
	"fmt"
	"strings"
	"testing"

	"golang.org/x/tools/go/packages"
)

const module = "github.com/JRamonCarralero/nexo/backend"

type rule struct {
	name  string
	from  func(pkg string) bool
	allow func(pkg, imp string) bool
}

var rules = []rule{
	{
		name: "a module does not import another module (talk through ports wired in app)",
		from: under("internal/modules"),
		allow: func(pkg, imp string) bool {
			return !under("internal/modules")(imp) || moduleOf(imp) == moduleOf(pkg)
		},
	},
	{
		name: "only app and cmd import concrete adapters (platform/<x>/<impl>)",
		from: func(pkg string) bool {
			return !under("internal/app")(pkg) && !under("cmd")(pkg) && !isTestHelper(pkg)
		},
		allow: func(pkg, imp string) bool {
			return !isAdapterImpl(imp) || strings.HasPrefix(pkg, parentOf(imp))
		},
	},
	{
		name:  "platform does not know the business",
		from:  under("internal/platform"),
		allow: deny("internal/modules", "internal/query", "internal/httpserver", "internal/app", "internal/bootstrap", "internal/devseed"),
	},
	{
		name:  "query is read-only and does not import modules",
		from:  under("internal/query"),
		allow: deny("internal/modules", "internal/app"),
	},
	{
		name:  "modules do not depend on wiring",
		from:  under("internal/modules"),
		allow: deny("internal/app", "internal/httpserver/middleware", "internal/bootstrap", "internal/devseed"),
	},
	{
		name:  "main stays thin: cmd does not import modules, query or HTTP internals",
		from:  under("cmd"),
		allow: deny("internal/modules", "internal/query", "internal/httpserver"),
	},
}

func TestArchitecture(t *testing.T) {
	pkgs, err := packages.Load(&packages.Config{Mode: packages.NeedName | packages.NeedImports}, "./...")
	if err != nil {
		t.Fatalf("loading packages: %v", err)
	}
	if len(pkgs) == 0 {
		t.Fatal("no packages loaded")
	}

	for _, p := range pkgs {
		for _, e := range p.Errors {
			t.Errorf("package %s: %v", p.PkgPath, e)
		}
		pkg := rel(p.PkgPath)
		for _, r := range rules {
			if !r.from(pkg) {
				continue
			}
			for impPath := range p.Imports {
				if !strings.HasPrefix(impPath, module+"/") {
					continue
				}
				if imp := rel(impPath); !r.allow(pkg, imp) {
					t.Errorf("%s\n\t%s imports %s", r.name, pkg, imp)
				}
			}
		}
	}
}

func TestRulesCatchViolations(t *testing.T) {
	cases := []struct{ pkg, imp string }{
		{"internal/modules/task", "internal/modules/project"},
		{"internal/modules/task", "internal/platform/auth/firebase"},
		{"internal/platform/database", "internal/modules/task"},
		{"internal/query", "internal/modules/task"},
		{"cmd/api", "internal/modules/task"},
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("%s->%s", c.pkg, c.imp), func(t *testing.T) {
			if violated(c.pkg, c.imp) == "" {
				t.Errorf("expected %s importing %s to violate a rule", c.pkg, c.imp)
			}
		})
	}

	allowed := []struct{ pkg, imp string }{
		{"internal/modules/task", "internal/platform/apperror"},
		{"internal/modules/task", "internal/platform/auth"},
		{"internal/modules/task", "internal/platform/database"},
		{"internal/modules/task/sub", "internal/modules/task"},
		{"internal/app", "internal/platform/auth/firebase"},
		{"internal/platform/auth/firebase", "internal/platform/auth"},
	}
	for _, c := range allowed {
		t.Run(fmt.Sprintf("allowed %s->%s", c.pkg, c.imp), func(t *testing.T) {
			if r := violated(c.pkg, c.imp); r != "" {
				t.Errorf("%s importing %s should be allowed, broke %q", c.pkg, c.imp, r)
			}
		})
	}
}

func violated(pkg, imp string) string {
	for _, r := range rules {
		if r.from(pkg) && !r.allow(pkg, imp) {
			return r.name
		}
	}
	return ""
}

func rel(path string) string { return strings.TrimPrefix(strings.TrimPrefix(path, module), "/") }

func under(prefix string) func(string) bool {
	return func(pkg string) bool { return pkg == prefix || strings.HasPrefix(pkg, prefix+"/") }
}

func deny(prefixes ...string) func(pkg, imp string) bool {
	return func(_, imp string) bool {
		for _, p := range prefixes {
			if under(p)(imp) {
				return false
			}
		}
		return true
	}
}

// moduleOf returns "internal/modules/<name>" for any package inside a module.
func moduleOf(pkg string) string {
	parts := strings.SplitN(pkg, "/", 4)
	if len(parts) < 3 {
		return pkg
	}
	return strings.Join(parts[:3], "/")
}

// isAdapterImpl reports whether pkg is a technology-specific implementation,
// i.e. a subpackage of a platform package (platform/auth/firebase). The
// database package is the exception: modules own their Postgres stores (D-008).
func isAdapterImpl(pkg string) bool {
	if !under("internal/platform")(pkg) || under("internal/platform/database")(pkg) {
		return false
	}
	return strings.Count(strings.TrimPrefix(pkg, "internal/platform/"), "/") >= 1
}

func parentOf(pkg string) string { return pkg[:strings.LastIndex(pkg, "/")] }

// isTestHelper exempts packages that only exist to support tests.
func isTestHelper(pkg string) bool { return strings.HasSuffix(pkg, "test") }
