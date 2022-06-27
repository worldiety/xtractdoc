package testdata

import other "github.com/worldiety/xtractdoc/testdata/v1"

// MyEnum is like Go does it.
type MyEnum string

// Group documentation on constants.
const (
	// AConstant here.
	AConstant MyEnum = "abc"

	// BConstant here.
	BConstant MyEnum = "bcd"
)

// An Entity to store.
type Entity struct {
	// A Name to tell about.
	Name string

	// A Description about the thing.
	Description map[string]int

	// GenericOne is recursive.
	GenericOne X[X[chan<- bool]]

	hidden string
}

type X[T any] struct {
	V T
}

// String returns a human-
// readable representation.
//
// Second line.
func (e Entity) String() string {
	return "hey"
}

// A Behavior is what to want.
type Behavior interface {
	// DoIt does it well.
	DoIt()
}

// Hello to the world.
var Hello = "world"

// grouped consts
const (
	// HelloConst as const
	HelloConst = "hello world"
)

// The BestFunc is really a static package level function.
func BestFunc() {

}

// NewEntity is a conventional constructor.
func NewEntity() Entity {
	return Entity{}
}

type MyFace interface {
	// A is virtual method.
	A() (other.DifferentV1, error)
}

type MyFaceImpl struct {
}

// A is a concrete method.
func (MyFaceImpl) A() (other.DifferentV1, error) {
	return other.DifferentV1{}, nil
}
