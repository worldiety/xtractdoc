package api

// Stereotype as usually interpreted in found context but not expressed in language explicitly.
type Stereotype string

const (
	StereotypeConstructor     = "constructor"
	StereotypeMethod          = "method"
	StereotypeSingleton       = "singleton"
	StereotypeEnum            = "enum"
	StereotypeEnumElement     = "enumElement"
	StereotypeDestructor      = "destructor"
	StereotypeExecutable      = "executable"
	StereotypeStruct          = "struct"
	StereotypeClass           = "class"
	StereotypeProperty        = "property"
	StereotypeParameter       = "parameter"
	StereotypeParameterIn     = "in"
	StereotypeParameterOut    = "out"
	StereotypeParameterResult = "result"
)

type ImportPath = string

type Module struct {
	Readme   string                  `json:"readme,omitempty" yaml:"readme,omitempty"`
	Module   string                  `json:"module" yaml:"module"`
	Packages map[ImportPath]*Package `json:"packages,omitempty" yaml:"packages,omitempty"`
}

type Package struct {
	Readme      string            `json:"readme,omitempty" yaml:"readme,omitempty"`
	Doc         string            `json:"doc,omitempty" yaml:"doc,omitempty"`
	Name        string            `json:"name" yaml:"name"`
	Imports     []string          `json:"imports,omitempty" yaml:"imports,omitempty"`
	Stereotypes []Stereotype      `json:"stereotypes,omitempty" yaml:"stereotypes,omitempty"`
	Types       map[string]*Type  `json:"types,omitempty" yaml:"types,omitempty"`
	Consts      map[string]*Const `json:"consts,omitempty" yaml:"consts,omitempty"`
	Vars        map[string]*Var   `json:"vars,omitempty" yaml:"vars,omitempty"`
	Functions   map[string]*Func  `json:"functions,omitempty" yaml:"functions,omitempty"`
}

type BaseType = string

type Type struct {
	Doc         string            `json:"doc,omitempty" yaml:"doc,omitempty"`
	BaseType    BaseType          `json:"baseType" yaml:"baseType"`
	Stereotypes []Stereotype      `json:"stereotypes,omitempty" yaml:"stereotypes,omitempty"`
	Factories   map[string]*Func  `json:"factories,omitempty" yaml:"factories,omitempty"`
	Methods     map[string]*Func  `json:"methods,omitempty" yaml:"methods,omitempty"`
	Singletons  map[string]*Var   `json:"singletons,omitempty" yaml:"singletons,omitempty"`
	Fields      map[string]*Field `json:"fields,omitempty" yaml:"fields,omitempty"`
	Enumerals   map[string]*Enum  `json:"enum,omitempty" yaml:"enum,omitempty"`
}

type Func struct {
	Doc         string                `json:"doc,omitempty" yaml:"doc,omitempty"`
	Stereotypes []Stereotype          `json:"stereotypes,omitempty" yaml:"stereotypes,omitempty"`
	Params      map[string]*Parameter `json:"params,omitempty" yaml:"params,omitempty"`
	Results     map[string]*Parameter `json:"results,omitempty" yaml:"results,omitempty"`
}

type Field struct {
	Doc         string       `json:"doc,omitempty" yaml:"doc,omitempty"`
	BaseType    BaseType     `json:"baseType" yaml:"baseType"`
	Stereotypes []Stereotype `json:"stereotypes,omitempty" yaml:"stereotypes,omitempty"`
}

type Parameter struct {
	Doc         string       `json:"doc,omitempty" yaml:"doc,omitempty"`
	BaseType    BaseType     `json:"baseType" yaml:"baseType"`
	Stereotypes []Stereotype `json:"stereotypes,omitempty" yaml:"stereotypes,omitempty"`
}

type Enum struct {
	Doc string `json:"doc,omitempty" yaml:"doc,omitempty"`
}

type Const struct {
	Doc         string       `json:"doc,omitempty" yaml:"doc,omitempty"`
	Stereotypes []Stereotype `json:"stereotypes,omitempty" yaml:"stereotypes,omitempty"`
}

type Var struct {
	Doc         string       `json:"doc,omitempty" yaml:"doc,omitempty"`
	Stereotypes []Stereotype `json:"stereotypes,omitempty" yaml:"stereotypes,omitempty"`
}
