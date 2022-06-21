package testdata

const (
	// AConstant here.
	AConstant = "abc"
)

// An Entity to store.
type Entity struct {
	// A Name to tell about.
	Name string

	// A Description about the thing.
	Description string
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

// The BestFunc is really a static package level function.
func BestFunc() {

}

// NewEntity is a conventional constructor.
func NewEntity() Entity {
	return Entity{}
}
