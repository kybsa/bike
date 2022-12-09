package bike

import (
	"reflect"
)

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
	instanceValue           *reflect.Value
	prototypeInstancesValue []*reflect.Value
}

// ErrorCode type to enum error codes
type ErrorCode uint8

const (
	// DependecyByIDNotFound error code when a dependency not found by id
	DependecyByIDNotFound ErrorCode = 0
	// DependecyByTypeNotFound error code when a dependency not found by type
	DependecyByTypeNotFound ErrorCode = 1
	// InvalidScope error code when a invalid scope
	InvalidScope ErrorCode = 2
	// InvalidNumArgOnPostConstruct error code when PostContruct method has arguments
	InvalidNumArgOnPostConstruct ErrorCode = 4
	// InvalidNumArgOnDestroy error code when Destroy method has arguments
	InvalidNumArgOnDestroy ErrorCode = 5
	// InvalidNumberOfReturnValuesOnConstructor error when contructor return invalid number of values
	InvalidNumberOfReturnValuesOnConstructor ErrorCode = 6
	// ConstructorReturnNoPointerValue error when constructor method return a no pointer value
	ConstructorReturnNoPointerValue ErrorCode = 7
	// ComponentConstructorNull error when Contructor property is null
	ComponentConstructorNull ErrorCode = 8
)

// Error structo with error info
type Error struct {
	messageError string
	errorCode    ErrorCode
}

// Error return message about error
func (_self *Error) Error() string {
	return _self.messageError
}

// ErrorCode return error code about error
func (_self *Error) ErrorCode() ErrorCode {
	return _self.errorCode
}

// Bike is main struct of this package
type Bike struct {
	componentsByType map[reflect.Type]*Component
	componentsByID   map[string]*Component
	components       []*Component
}

// Container struct with component management
type Container struct {
	bike *Bike
}

// NewBike create a Bike instance
func NewBike() *Bike {
	return &Bike{
		componentsByType: make(map[reflect.Type]*Component),
		componentsByID:   make(map[string]*Component),
		components:       make([]*Component, 0),
	}
}

// Registry add component to bike
func (_self *Bike) Registry(component Component) *Error {

	// Check if component have not constructor method
	if component.Constructor == nil {
		return &Error{messageError: "Constructor must no be nill", errorCode: ComponentConstructorNull}
	}

	// Check Constructor component
	constructorType := reflect.TypeOf(component.Constructor)
	if constructorType.NumOut() != 1 {
		return &Error{messageError: "Constructor must return one value", errorCode: InvalidNumberOfReturnValuesOnConstructor}
	}
	typeComponent := constructorType.Out(0)

	// Check return Constructor value
	if typeComponent.Kind() != reflect.Pointer && typeComponent.Kind() != reflect.Interface {
		return &Error{messageError: "Constructor must return a pointer o interface value", errorCode: ConstructorReturnNoPointerValue}
	}

	// Check PostConstruct
	if len([]rune(component.PostConstruct)) > 0 {
		componentType := constructorType.Out(0)
		method, ok := componentType.MethodByName(component.PostConstruct)
		if !ok {
			return &Error{messageError: "Component.PostConstruct [" + component.PostConstruct + "] not found" + component.PostConstruct, errorCode: InvalidNumArgOnPostConstruct}
		}
		if method.Type.NumIn() != 1 {
			return &Error{messageError: "Invalid argument number of Component.PostConstruct. PostConstruct:" + component.PostConstruct, errorCode: InvalidNumArgOnPostConstruct}
		}
	}

	// Check Destroy
	if len([]rune(component.Destroy)) > 0 {
		componentType := constructorType.Out(0)
		method, ok := componentType.MethodByName(component.Destroy)
		if !ok {
			return &Error{messageError: "Invalid Component.Destroy:" + component.Destroy, errorCode: InvalidNumArgOnPostConstruct}
		}
		if method.Type.NumIn() != 1 {
			return &Error{messageError: "Invalid number arguments of Component.Destroy:" + component.Destroy, errorCode: InvalidNumArgOnPostConstruct}
		}
	}

	// Registry by id
	if len([]rune(component.ID)) > 0 {
		_self.componentsByID[component.ID] = &component
	}

	_self.componentsByType[typeComponent] = &component

	// Registry by interfaces
	for _, inter := range component.Interfaces {
		interfaceType := reflect.TypeOf(inter).Elem()
		_self.componentsByType[interfaceType] = &component
	}

	// Init array of prototype instances
	if component.Scope == Prototype {
		component.prototypeInstancesValue = make([]*reflect.Value, 0)
	} else if component.Scope != Singleton {
		message := "Invalid Scope: " + component.Scope.String()
		return &Error{messageError: message, errorCode: InvalidScope}
	}

	// Add to array
	_self.components = append(_self.components, &component)

	return nil
}

func (_self *Bike) instanceByTypeAny(inputType any) (interface{}, *Error) {
	_type := reflect.TypeOf(inputType)
	if _type.Kind() == reflect.Pointer && _type.Elem().Kind() == reflect.Interface {
		_type = _type.Elem()
	}
	return _self.instanceByType(_type)
}

func (_self *Bike) instanceByID(id string) (interface{}, *Error) {
	component, ok := _self.componentsByID[id]
	if ok {
		if component.Scope == Singleton {
			if component.instanceValue.Elem().CanAddr() {
				return component.instanceValue.Elem().Addr().Interface(), nil
			}
			return component.instanceValue.Elem().Interface(), nil
		} else if component.Scope == Prototype {
			instance, err := _self.createComponent(component)
			if err != nil {
				return nil, err
			}
			var interfaceInstance any
			if instance.Elem().CanAddr() {
				interfaceInstance = instance.Elem().Addr().Interface()
			} else {
				interfaceInstance = instance.Elem().Interface()
			}
			component.prototypeInstancesValue = append(component.prototypeInstancesValue, instance)
			return interfaceInstance, nil
		}
	}
	message := "Component by id:" + id + " not found"
	return nil, &Error{messageError: message, errorCode: DependecyByIDNotFound}
}

func (_self *Bike) instanceByType(_type reflect.Type) (interface{}, *Error) {
	component, ok := _self.componentsByType[_type]
	if ok {
		if component.Scope == Singleton {
			return component.instanceValue.Elem().Addr().Interface(), nil
		} else if component.Scope == Prototype {
			instance, err := _self.createComponent(component)
			if err != nil {
				return nil, err
			}
			var interfaceInstance any
			if instance.Elem().CanAddr() {
				interfaceInstance = instance.Elem().Addr().Interface()
			} else {
				interfaceInstance = instance.Elem().Interface()
			}
			component.prototypeInstancesValue = append(component.prototypeInstancesValue, instance)
			return interfaceInstance, nil
		}
	}
	var message string
	if _type.Kind() == reflect.Interface {
		message = "Component by type:" + _type.Name() + " not found"

	} else {
		message = "Component by type:" + _type.Elem().Name() + " not found"
	}
	return nil, &Error{messageError: message, errorCode: DependecyByTypeNotFound}
}

func (_self *Bike) createComponent(component *Component) (*reflect.Value, *Error) {
	// Create component by contructor method
	constructorValue := reflect.ValueOf(component.Constructor)
	constructorType := reflect.TypeOf(component.Constructor)

	// Search dependecies
	args := make([]reflect.Value, constructorType.NumIn())
	for i := 0; i < constructorType.NumIn(); i++ {
		inputType := constructorType.In(i)
		inputArg, err := _self.instanceByType(inputType)
		if err == nil {
			args[i] = reflect.ValueOf(inputArg)
		} else {
			message := "Error to get dependecy: [" + inputType.Name() + "] required by function" + constructorType.Name() + "]\n" + err.Error()
			return nil, &Error{messageError: message, errorCode: err.ErrorCode()}
		}
	}
	// Create component with dependencies
	instanceResult := constructorValue.Call(args)
	component.instanceValue = &instanceResult[0]

	// Call init methods
	// Search Components methods to pointer struct
	componentType := constructorType.Out(0)
	if len([]rune(component.PostConstruct)) > 0 {
		method, _ := componentType.MethodByName(component.PostConstruct)
		method.Func.Call([]reflect.Value{*component.instanceValue})
	}

	return component.instanceValue, nil
}

// Start start bike
func (_self *Bike) Start() (*Container, *Error) {
	for _, component := range _self.components {
		// Create components
		if component.Scope == Singleton {
			instanceValue, err := _self.createComponent(component)
			if err != nil {
				return nil, err
			}
			component.instanceValue = instanceValue
		}
	}
	return &Container{bike: _self}, nil
}

// Stop stop bike
func (_self *Bike) Stop() *Error {
	var lastError *Error
	for _, component := range _self.components {
		if len([]rune(component.Destroy)) > 0 {
			componentType := reflect.TypeOf(component.Constructor).Out(0)
			method, _ := componentType.MethodByName(component.Destroy)
			if component.Scope == Singleton {
				method.Func.Call([]reflect.Value{*component.instanceValue})
			} else if component.Scope == Prototype {
				for _, prototypeInstance := range component.prototypeInstancesValue {
					method.Func.Call([]reflect.Value{*prototypeInstance})
				}
			}
		}
	}
	return lastError
}

// InstanceByType return a instance by type
func (_self *Container) InstanceByType(inputType any) (interface{}, *Error) {
	return _self.bike.instanceByTypeAny(inputType)
}

// InstanceByID return a instance by ID
func (_self *Container) InstanceByID(id string) (interface{}, *Error) {
	return _self.bike.instanceByID(id)
}
