package golang

import (
	"fmt"
	"github.com/worldiety/xtractdoc/internal/api"
	"go/ast"
	"go/doc"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func newModule(dir string, modname string, pkgs map[string]Package) (*api.Module, error) {
	m := &api.Module{
		Module: modname,
		Readme: tryLoadReadme(dir),
	}

	if len(pkgs) > 0 {
		m.Packages = map[api.ImportPath]*api.Package{}
		for _, p := range pkgs {
			np := newPackage(p)
			np.Readme = tryLoadReadme(p.dir)
			m.Packages[p.dpkg.ImportPath] = np

		}
	}

	return m, nil
}

func tryLoadReadme(dir string) string {
	files, _ := os.ReadDir(dir)
	for _, file := range files {
		if file.Type().IsRegular() && strings.ToLower(file.Name()) == "readme.md" {
			buf, _ := os.ReadFile(filepath.Join(dir, file.Name()))
			if len(buf) != 0 {
				return string(buf)
			}
		}
	}

	return ""
}

func newPackage(pkg Package) *api.Package {
	p := &api.Package{
		Doc:     pkg.dpkg.Doc,
		Name:    pkg.dpkg.Name,
		Imports: pkg.dpkg.Imports,
	}

	if pkg.dpkg.Name == "main" {
		p.Stereotypes = append(p.Stereotypes, api.StereotypeExecutable)
	}

	if len(pkg.dpkg.Funcs) > 0 {
		p.Functions = map[string]*api.Func{}
		for _, f := range pkg.dpkg.Funcs {
			if !f.Decl.Name.IsExported() {
				continue
			}

			p.Functions[f.Name] = newFunc(f)
		}

	}

	if len(pkg.dpkg.Types) > 0 {
		p.Types = map[string]*api.Type{}
		for _, t := range pkg.dpkg.Types {
			p.Types[t.Name] = newType(t)
		}
	}

	if len(pkg.dpkg.Consts) > 0 {
		p.Consts = map[string]*api.Const{}
		for _, value := range pkg.dpkg.Consts {
			for _, d := range newValue(value) {
				p.Consts[d.name] = &api.Const{Doc: d.doc}
			}
		}
	}

	if len(pkg.dpkg.Vars) > 0 {
		p.Vars = map[string]*api.Var{}
		for _, value := range pkg.dpkg.Vars {
			for _, d := range newValue(value) {
				p.Vars[d.name] = &api.Var{Doc: d.doc}
			}
		}
	}

	return p
}

func newFunc(fn *doc.Func) *api.Func {
	f := &api.Func{
		Doc: fn.Doc,
	}

	inArgs := fn.Decl.Type.Params.List
	if len(inArgs) > 0 {
		f.Params = map[string]*api.Parameter{}
		insertParams(f.Params, inArgs, api.StereotypeParameter, api.StereotypeParameterIn)
	}

	if fn.Decl.Type.Results != nil {
		outArgs := fn.Decl.Type.Results.List
		if len(outArgs) > 0 {
			f.Results = map[string]*api.Parameter{}
			insertParams(f.Results, outArgs, api.StereotypeParameter, api.StereotypeParameterOut, api.StereotypeParameterResult)
		}
	}

	return f
}

func insertParams(dst map[string]*api.Parameter, src []*ast.Field, st ...api.Stereotype) {
	c := 0
	for fnum, field := range src {
		if len(field.Names) == 0 {
			in := newField(field)
			dst["__"+strconv.Itoa(fnum)] = &api.Parameter{
				Doc:         in.Doc,
				BaseType:    in.BaseType,
				Stereotypes: st,
			}
			continue
		}

		for _, name := range field.Names {
			c++
			in := newField(field)
			myName := name.Name
			if myName == "" {
				myName = "__" + strconv.Itoa(c)
			}
			dst[name.Name] = &api.Parameter{
				Doc:         in.Doc,
				BaseType:    in.BaseType,
				Stereotypes: st,
			}
		}
	}
}

func newType(typeDef *doc.Type) *api.Type {
	n := &api.Type{
		Doc: typeDef.Doc,
	}

	for _, spec := range typeDef.Decl.Specs {
		n.BaseType = ast2str(spec)
		switch t := spec.(type) {
		case *ast.TypeSpec:
			switch t := t.Type.(type) {
			case *ast.StructType:
				n.Stereotypes = append(n.Stereotypes, api.StereotypeStruct, api.StereotypeClass)
				if len(t.Fields.List) > 0 {
					n.Fields = map[string]*api.Field{}
					for _, field := range t.Fields.List {
						nf := newField(field)
						for _, name := range field.Names {
							if !name.IsExported() {
								continue
							}

							n.Fields[name.Name] = nf
						}
					}
				}
			case *ast.Ident:
				n.Stereotypes = append(n.Stereotypes, api.Stereotype(t.Name))
			}
		}
	}

	if len(typeDef.Consts) > 0 {
		n.Stereotypes = append(n.Stereotypes, api.StereotypeEnum)
		n.Enumerals = map[string]*api.Enum{}
	}

	for _, value := range typeDef.Consts {
		for _, d := range newValue(value) {
			enum := &api.Enum{
				Doc: d.doc,
			}
			n.Enumerals[d.name] = enum
		}
	}

	if len(typeDef.Funcs) > 0 {
		n.Factories = map[string]*api.Func{}
		for _, f := range typeDef.Funcs {
			nf := newFunc(f)
			nf.Stereotypes = append(nf.Stereotypes, api.StereotypeConstructor)
			n.Factories[f.Name] = nf
		}

	}

	if len(typeDef.Methods) > 0 {
		n.Methods = map[string]*api.Func{}
		for _, f := range typeDef.Methods {
			nf := newFunc(f)
			nf.Stereotypes = append(nf.Stereotypes, api.StereotypeMethod)
			n.Methods[f.Name] = nf
		}

	}

	if len(typeDef.Vars) > 0 {
		n.Singletons = map[string]*api.Var{}
		for _, value := range typeDef.Vars {
			for _, d := range newValue(value) {
				n.Singletons[d.name] = &api.Var{Doc: d.doc, Stereotypes: []api.Stereotype{api.StereotypeSingleton}}
			}
		}
	}

	return n
}

func newField(field *ast.Field) *api.Field {
	n := &api.Field{
		Doc:         field.Doc.Text(),
		BaseType:    ast2str(field.Type),
		Stereotypes: []api.Stereotype{api.StereotypeProperty},
	}

	return n
}

type docValue struct {
	doc  string
	name string
}

func newValue(value *doc.Value) []docValue {
	var res []docValue
	groupDoc := value.Doc
	for _, spec := range value.Decl.Specs {
		switch t := spec.(type) {
		case *ast.ValueSpec:
			actualDoc := t.Doc.Text()
			for _, name := range t.Names {
				if !name.IsExported() {
					continue
				}
				res = append(res, docValue{
					doc:  strings.TrimSpace(groupDoc + "\n" + actualDoc),
					name: name.Name,
				})
			}

		}
	}

	return res
}

func ast2str(n ast.Node) string {
	switch t := n.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return ast2str(t.X) + "." + t.Sel.Name
	case *ast.StarExpr:
		return "*" + ast2str(t.X)
	case *ast.MapType:
		return "map[" + ast2str(t.Key) + "]" + ast2str(t.Value)
	case *ast.IndexExpr:
		return ast2str(t.X) + "[" + ast2str(t.Index) + "]"
	case *ast.ChanType:
		s := "chan"
		switch t.Dir {
		case ast.SEND:
			s += "<-"
		case ast.RECV:
			s += "->"
		}

		s += ast2str(t.Value)

		return s
	case *ast.ArrayType:
		s := "["
		if t.Len != nil {
			s += ast2str(t.Len)
		}
		s += "]"
		s += ast2str(t.Elt)
		return s
	case *ast.TypeSpec:
		return ast2str(t.Type)
	case *ast.InterfaceType:
		return "interface"
	case *ast.StructType:
		return "struct"
	case *ast.Ellipsis:
		return "..."
	default:
		panic(fmt.Errorf("implement me %T", t))
	}
}
