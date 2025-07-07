package golang

import (
	"bytes"
	"fmt"
	"github.com/worldiety/xtractdoc/domain/api"
	"go/ast"
	"go/doc"
	"go/format"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func newModule(dir string, modname string, pkgs map[string]Package, fset *token.FileSet) (*api.Module, error) {
	m := &api.Module{
		Module:   modname,
		Readme:   tryLoadReadme(dir),
		Packages: make(map[api.ImportPath]*api.Package),
	}

	for _, p := range pkgs {
		var exampleFuncs map[string][]*ast.FuncDecl
		var exampleImports map[string][][]*ast.ImportSpec

		// find all associated _test-Package
		testPkg, ok := pkgs[p.pkg.Name+"_test"]
		if ok {
			exampleFuncs = make(map[string][]*ast.FuncDecl)
			exampleImports = make(map[string][][]*ast.ImportSpec)

			for _, f := range testPkg.pkg.Files {
				for _, decl := range f.Decls {
					fn, ok := decl.(*ast.FuncDecl)
					if !ok || !strings.HasPrefix(fn.Name.Name, "Example") {
						continue
					}

					base := exampleTargetFunc(fn.Name.Name)
					exampleFuncs[base] = append(exampleFuncs[base], fn)
					exampleImports[base] = append(exampleImports[base], f.Imports)
				}
			}
		}

		np := newPackage(p, fset, exampleFuncs, exampleImports)
		np.Readme = tryLoadReadme(p.dir)
		m.Packages[p.dpkg.ImportPath] = np
	}

	return m, nil
}

func exampleTargetFunc(name string) string {
	s := strings.TrimPrefix(name, "Example")
	if i := strings.Index(s, "_"); i != -1 {
		return s[:i]
	}
	return s
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

func newPackage(pkg Package, fset *token.FileSet, exampleFuncs map[string][]*ast.FuncDecl, exampleImports map[string][][]*ast.ImportSpec) *api.Package {
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

			p.Functions[f.Name] = newFuncFromDoc(f, fset, exampleFuncs, exampleImports)
		}

	}

	if len(pkg.dpkg.Types) > 0 {
		p.Types = map[string]*api.Type{}
		for _, t := range pkg.dpkg.Types {
			p.Types[t.Name] = newType(t, fset, exampleFuncs, exampleImports)
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

func newFuncFromDoc(df *doc.Func, fset *token.FileSet, testExamples map[string][]*ast.FuncDecl, testImports map[string][][]*ast.ImportSpec) *api.Func {
	f := &api.Func{
		Doc: df.Doc,
	}

	inArgs := df.Decl.Type.Params.List
	if len(inArgs) > 0 {
		f.Params = map[string]*api.Parameter{}
		insertParams(f.Params, inArgs, api.StereotypeParameter, api.StereotypeParameterIn)
	}

	if df.Decl.Type.Results != nil {
		outArgs := df.Decl.Type.Results.List
		if len(outArgs) > 0 {
			f.Results = map[string]*api.Parameter{}
			insertParams(f.Results, outArgs, api.StereotypeParameter, api.StereotypeParameterOut, api.StereotypeParameterResult)
		}
	}

	for _, ex := range df.Examples {
		f.Examples = append(f.Examples, api.Example{
			Name: ex.Name,
			Doc:  ex.Doc,
			Code: exampleCodeToString(fset, ex.Code),
		})
	}

	if exFns, ok := testExamples[df.Name]; ok {
		importGroups := testImports[df.Name]
		for i, exFn := range exFns {
			var imports []*ast.ImportSpec
			if i < len(importGroups) {
				imports = importGroups[i]
			}
			example, err := generateExecutableFromExample(exFn, imports, fset)
			if err == nil {
				f.ExecutableExamples = append(f.ExecutableExamples, example)
			}
		}
	}

	return f
}

func exampleCodeToString(fset *token.FileSet, node ast.Node) string {
	if node == nil {
		return ""
	}

	var buf bytes.Buffer
	err := printer.Fprint(&buf, fset, node)
	if err != nil {
		return ""
	}

	formatted, err := format.Source(buf.Bytes())
	if err == nil {
		return string(formatted)
	}
	return buf.String()
}

func newFuncFromAst(doc string, fn *ast.FuncType) *api.Func {
	f := &api.Func{
		Doc: doc,
	}

	if fn.Params != nil {
		inArgs := fn.Params.List
		if len(inArgs) > 0 {
			f.Params = map[string]*api.Parameter{}
			insertParams(f.Params, inArgs, api.StereotypeParameter, api.StereotypeParameterIn)
		}
	}

	if fn.Results != nil {
		outArgs := fn.Results.List
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

func newType(typeDef *doc.Type, fset *token.FileSet, exampleFuncs map[string][]*ast.FuncDecl, exampleImports map[string][][]*ast.ImportSpec) *api.Type {
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
			case *ast.InterfaceType:
				n.Methods = map[string]*api.Func{}
				for _, f := range t.Methods.List {
					switch m := f.Type.(type) {
					case *ast.FuncType:
						nf := newFuncFromAst(f.Doc.Text(), m)
						nf.Stereotypes = append(nf.Stereotypes, api.StereotypeMethod)
						for _, name := range f.Names {
							if !name.IsExported() {
								continue
							}
							n.Methods[name.Name] = nf
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
			nf := newFuncFromDoc(f, fset, exampleFuncs, exampleImports)
			n.Factories[f.Name] = nf
			n.Factories[f.Name] = nf
		}

	}

	if len(typeDef.Methods) > 0 {
		n.Methods = map[string]*api.Func{}
		for _, f := range typeDef.Methods {
			nf := newFuncFromDoc(f, fset, exampleFuncs, exampleImports)
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
	case *ast.FuncType:
		s := "func"
		if t.TypeParams != nil && len(t.TypeParams.List) > 0 {
			s += "["
			s += ast2str(t.TypeParams)
			s += "]"
		}
		s += "("
		if t.Params != nil {
			s += ast2str(t.Params)
		}
		s += ")"
		if t.Results != nil {
			switch len(t.Results.List) {
			case 0:
			case 1:
				s += " "
				s += ast2str(t.Results)
			default:
				s += " ("
				s += ast2str(t.Results)
				s += ")"
			}
		}
		return s
	case *ast.FieldList:
		s := ""
		for _, field := range t.List {
			for _, name := range field.Names {
				s += name.Name + " ,"
			}
			s = strings.TrimSuffix(s, ",")
			s += ast2str(field.Type)
		}

		s = strings.TrimSuffix(s, " ,")
		return s
	default:
		panic(fmt.Errorf("implement me %T", t))
	}
}

// generateExecutableFromExample generates executable code from an example with a main package,
// all required imports and the code wrapped into a main function.
//
// e.g.
//
//	ExampleSayHello() {
//	 fmt.Println("Hello")
//	}
//
// result:
//
// package main
//
// import "fmt"
//
//	func main() {
//	  fmt.Println("Hello")
//	}
func generateExecutableFromExample(fn *ast.FuncDecl, imports []*ast.ImportSpec, fset *token.FileSet) (api.ExecutableExample, error) {
	filteredImports := filterUsedImports(fn, imports)

	mainFunc := &ast.FuncDecl{
		Name: ast.NewIdent("main"),
		Type: &ast.FuncType{
			Params: &ast.FieldList{},
		},
		Body: fn.Body,
	}

	importDecl := &ast.GenDecl{
		Tok:   token.IMPORT,
		Specs: make([]ast.Spec, len(filteredImports)),
	}

	for i, imp := range filteredImports {
		importDecl.Specs[i] = imp
	}

	file := &ast.File{
		Name:  mainFunc.Name,
		Decls: []ast.Decl{importDecl, mainFunc},
	}

	var buf bytes.Buffer

	err := format.Node(&buf, fset, file)
	if err != nil {
		return api.ExecutableExample{}, err
	}

	return api.ExecutableExample{
		Code: buf.String(),
	}, nil
}

func filterUsedImports(fn *ast.FuncDecl, imports []*ast.ImportSpec) []*ast.ImportSpec {
	// 1. find all SelectorExpr like fmt.Println, strings.NewReader
	usedIdents := map[string]bool{}

	ast.Inspect(fn.Body, func(n ast.Node) bool {
		sel, ok := n.(*ast.SelectorExpr)
		if ok {
			if ident, ok := sel.X.(*ast.Ident); ok {
				usedIdents[ident.Name] = true
			}
		}
		return true
	})

	// 2. filter all imports that are included in usedIdents
	var filtered []*ast.ImportSpec

	for _, imp := range imports {
		var name string

		if imp.Name != nil {
			name = imp.Name.Name // alias: z. B. "io"
		} else {
			// fallback: extract path → z. B. "fmt"
			path, err := strconv.Unquote(imp.Path.Value)
			if err != nil {
				continue
			}
			parts := strings.Split(path, "/")
			name = parts[len(parts)-1]
		}

		if usedIdents[name] {
			filtered = append(filtered, imp)
		}
	}

	return filtered
}
