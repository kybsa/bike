package bike

import "reflect"

// Scope Component supported
type Scope uint8

const (
	// Singleton scope, one instance to bike instance
	Singleton Scope = 0
	// Prototype scope, n instances to bike instance
	Prototype Scope = 1
)

var mapScopeString = map[Scope]string{
	Singleton: "Singleton",
	Prototype: "Prototype",
}

func (_self *Scope) String() string {
	return mapScopeString[*_self]
}

// Component is a struct with data to create components
type Component struct {
	ID                      string
	Interfaces              []any
	Scope                   Scope
	PostConstruct           string
	Destroy                 string
	Constructor             interface{}
	PostStart               string
	instanceValue           *reflect.Value
	prototypeInstancesValue []*reflect.Value
}
