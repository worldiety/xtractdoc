package golang

import (
	"fmt"
	"github.com/worldiety/xtractdoc/internal/api"
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"golang.org/x/exp/slices"
	"io/fs"
	"path/filepath"
	"strings"
)

type Package struct {
	pkg  *ast.Package
	dpkg *doc.Package
	dir  string
}

func Parse(dir string, onlyImports ...string) (*api.Module, error) {
	modRoot, err := ModRoot(dir)
	if err != nil {
		return nil, fmt.Errorf("cannot detect module root: %w", err)
	}

	modName, err := ModulePath(dir)
	if err != nil {
		return nil, fmt.Errorf("cannot detect go module path: %w", err)
	}

	dirs, err := PkgDirs(modRoot)
	if err != nil {
		return nil, fmt.Errorf("cannot determine pkg dirs: %w", err)
	}

	fset := token.NewFileSet()

	module := map[string]Package{} // import-path => parsed comments
	for _, dir := range dirs {
		rel, err := filepath.Rel(modRoot, dir)
		if err != nil {
			panic(fmt.Errorf("cannot happen: %w", err))
		}

		pkgs, err := parser.ParseDir(fset, dir, func(info fs.FileInfo) bool {
			return strings.HasSuffix(info.Name(), ".go")
		}, parser.ParseComments)

		if err != nil {
			return nil, fmt.Errorf("cannot parse: %w", err)
		}

		importPath := modName + "/" + rel
		if len(onlyImports) > 0 {
			if !slices.Contains(onlyImports, importPath) {
				continue
			}
		}

		for _, astPkg := range pkgs {
			pkg := doc.New(astPkg, importPath, doc.AllDecls)
			module[pkg.ImportPath] = Package{
				pkg:  astPkg,
				dpkg: pkg,
				dir:  dir,
			}
		}
	}

	return newModule(modRoot, modName, module)
}
