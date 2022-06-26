package golang

import (
	"fmt"
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"golang.org/x/exp/slices"
	"golang.org/x/mod/modfile"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func GetFields(t *doc.Type) []*ast.Field {
	for _, spec := range t.Decl.Specs {
		if tspec, ok := spec.(*ast.TypeSpec); ok {
			switch t := tspec.Type.(type) {
			case *ast.StructType:
				return t.Fields.List
			case *ast.InterfaceType:
				return t.Methods.List
			}

		}
	}

	return nil
}

// PkgDirs returns all directories containing any go file.
func PkgDirs(root string) ([]string, error) {
	var res []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() && strings.HasPrefix(d.Name(), ".") {
			return filepath.SkipDir
		}

		if d.IsDir() {
			files, err := os.ReadDir(path)
			if err != nil {
				return err
			}

			anyGoFile := false
			for _, file := range files {
				if file.Type().IsRegular() && strings.HasSuffix(file.Name(), ".go") {
					anyGoFile = true
					break
				}
			}

			if anyGoFile {
				res = append(res, path)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return res, nil
}

// ModWdRoot walks up from current working dir to find the enclosing go module.
func ModWdRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("cannot determine working dir: %w", err)
	}

	return ModRoot(cwd)
}

func ModRoot(cwd string) (string, error) {
	for cwd != "." && cwd != "/" && cwd != "" {
		modFile := filepath.Join(cwd, "go.mod")
		if _, err := os.Stat(modFile); err == nil {
			return cwd, nil
		}

		cwd = filepath.Dir(cwd)
	}

	return "", fmt.Errorf("no go.mod found")
}

// ModulePath returns whatever the mod path is.
func ModulePath(dir string) (string, error) {
	mdir, err := ModRoot(dir)
	if err != nil {
		return "", err
	}

	mf := filepath.Join(mdir, "go.mod")
	buf, err := os.ReadFile(mf)
	if err != nil {
		return "", fmt.Errorf("cannot open %s: %w", mf, err)
	}

	return modfile.ModulePath(buf), nil
}

type Type string

const (
	Package         Type = "package"
	TypeDef         Type = "typedef"
	Field           Type = "field"
	Func            Type = "func"
	Const           Type = "const"
	Var             Type = "var"
	StructMethod    Type = "receiver"
	InterfaceMethod Type = "method"
	Constructor     Type = "constructor"
	EnumConst       Type = "enumval"
)

type Comment struct {
	Qualifier string
	Type      Type
	Doc       string
}

func Parse(dir string, onlyImports ...string) ([]Comment, error) {
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
	var res []Comment

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
			if pkg.Doc != "" {
				res = append(res, Comment{
					Qualifier: importPath,
					Type:      Package,
					Doc:       pkg.Doc,
				})
			}

			for _, t := range pkg.Types {
				if t.Doc != "" {
					typeKey := importPath + "." + t.Name
					res = append(res, Comment{
						Qualifier: typeKey,
						Type:      TypeDef,
						Doc:       t.Doc,
					})
				}

				fields := GetFields(t)
				for _, field := range fields {
					if s := field.Doc.Text(); s != "" {
						for _, name := range field.Names {
							fieldKey := importPath + "." + t.Name + "#" + name.Name
							res = append(res, Comment{
								Qualifier: fieldKey,
								Type:      Field,
								Doc:       s,
							})
						}
					}
				}

				for _, method := range t.Methods {
					if method.Doc == "" {
						continue
					}

					fieldKey := importPath + "." + t.Name + "#" + method.Name
					res = append(res, Comment{
						Qualifier: fieldKey,
						Type:      StructMethod,
						Doc:       method.Doc,
					})
				}

				for _, f := range t.Funcs {
					if f.Doc == "" {
						continue
					}

					fieldKey := importPath + "." + t.Name + "#" + f.Name
					res = append(res, Comment{
						Qualifier: fieldKey,
						Type:      Constructor,
						Doc:       f.Doc,
					})
				}

				for _, c := range t.Consts {
					for _, spec := range c.Decl.Specs {
						if vspec, ok := spec.(*ast.ValueSpec); ok {
							for _, name := range vspec.Names {
								typeKey := importPath + "." + name.Name
								res = append(res, Comment{
									Qualifier: typeKey,
									Type:      EnumConst,
									Doc:       vspec.Doc.Text(),
								})
							}
						}
					}
				}
			}

			for _, f := range pkg.Funcs {
				if f.Doc != "" {
					typeKey := importPath + "." + f.Name
					res = append(res, Comment{
						Qualifier: typeKey,
						Type:      Func,
						Doc:       f.Doc,
					})
				}
			}

			for _, c := range pkg.Consts {
				if doc := valueDoc(c); doc != "" {
					for _, name := range c.Names {
						typeKey := importPath + "." + name
						res = append(res, Comment{
							Qualifier: typeKey,
							Type:      Const,
							Doc:       doc,
						})
					}
				}
			}

			for _, c := range pkg.Vars {
				if doc := valueDoc(c); doc != "" {
					for _, name := range c.Names {
						typeKey := importPath + "." + name
						res = append(res, Comment{
							Qualifier: typeKey,
							Type:      Var,
							Doc:       doc,
						})
					}

				}
			}
		}
	}

	return res, nil
}

func valueDoc(e *doc.Value) string {
	var sb strings.Builder
	sb.WriteString(e.Doc)
	for _, spec := range e.Decl.Specs {
		if vspec, ok := spec.(*ast.ValueSpec); ok {
			sb.WriteString(vspec.Doc.Text())
		}
	}

	return sb.String()

}
