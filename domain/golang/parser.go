package golang

import (
	"fmt"
	"github.com/worldiety/xtractdoc/domain/api"
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"golang.org/x/exp/slices"
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
	module := map[string]Package{}

	for _, dir := range dirs {
		rel, err := filepath.Rel(modRoot, dir)
		if err != nil {
			panic(fmt.Errorf("cannot happen: %w", err))
		}

		pkgs, err := parser.ParseDir(fset, dir, nil, parser.ParseComments)
		if err != nil {
			return nil, fmt.Errorf("cannot parse: %w", err)
		}

		importPath := modName + "/" + rel

		if len(onlyImports) > 0 {
			if !slices.Contains(onlyImports, importPath) {
				continue
			}
		}

		//var realPkgName string
		//for pkgName := range pkgs {
		//	if !strings.HasSuffix(pkgName, "_test") {
		//		realPkgName = pkgName
		//		break
		//	}
		//}
		mainPkg, testPkg := splitTestPackages(pkgs)

		if mainPkg != nil {
			module[mainPkg.Name] = buildPackageEntry(mainPkg, importPath, dir, fset)
		}

		if testPkg != nil {
			module[testPkg.Name] = buildPackageEntry(testPkg, "", dir, fset)
		}

		//for pkgName, astPkg := range pkgs {
		//	var files []*ast.File
		//	for _, f := range astPkg.Files {
		//		files = append(files, f)
		//	}
		//
		//	if pkgName == realPkgName {
		//		dpkg, err := doc.NewFromFiles(fset, files, importPath, doc.AllDecls|doc.AllMethods)
		//		if err != nil {
		//			panic(fmt.Errorf("cannot parse %s: %w", importPath, err))
		//		}
		//		module[pkgName] = Package{
		//			pkg:  astPkg,
		//			dpkg: dpkg,
		//			dir:  dir,
		//		}
		//	} else if strings.HasSuffix(pkgName, "_test") {
		//		dpkg, err := doc.NewFromFiles(fset, files, "", doc.AllDecls|doc.AllMethods)
		//		if err != nil {
		//			panic(fmt.Errorf("cannot parse %s: %w", importPath, err))
		//		}
		//		module[pkgName] = Package{
		//			pkg:  astPkg,
		//			dpkg: dpkg,
		//			dir:  dir,
		//		}
		//	}
		//}
	}

	return newModule(modRoot, modName, module, fset)
}

func splitTestPackages(pkgs map[string]*ast.Package) (*ast.Package, *ast.Package) {
	var mainPkg *ast.Package
	var testPkg *ast.Package

	for name, pkg := range pkgs {
		if strings.HasSuffix(name, "_test") {
			testPkg = pkg
		} else {
			mainPkg = pkg
		}
	}
	return mainPkg, testPkg
}

func buildPackageEntry(astPkg *ast.Package, importPath, dir string, fset *token.FileSet) Package {
	var files []*ast.File

	for _, f := range astPkg.Files {
		files = append(files, f)
	}

	dpkg, err := doc.NewFromFiles(fset, files, importPath, doc.AllDecls|doc.AllMethods)
	if err != nil {
		panic(fmt.Errorf("cannot parse %s: %w", importPath, err))
	}

	return Package{
		pkg:  astPkg,
		dpkg: dpkg,
		dir:  dir,
	}
}
