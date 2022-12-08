package bike

import (
	"reflect"
)

type Scope uint8

const (
	Singleton Scope = 0
	Prototype Scope = 1
)

var mapScopeString = map[Scope]string{
	Singleton: "Singleton",
	Prototype: "Prototype",
}

func (_self *Scope) String() string {
	return mapScopeString[*_self]
}

type Dependency struct {
	Type any
	Id   string
}

type Component struct {
	Id                      string
	Interfaces              []any
	Scope                   Scope
	Dependencies            map[string]Dependency
	PostConstruct           string
	Destroy                 string
	Constructor             interface{}
	instanceValue           *reflect.Value
	prototypeInstancesValue []*reflect.Value
}

type BikeErrorCode uint8

const (
	DependecyByIdNotFound                BikeErrorCode = 0
	DependecyByTypeNotFound              BikeErrorCode = 1
	InvalidScope                         BikeErrorCode = 2
	ComponentNotFound                    BikeErrorCode = 3
	InvalidNumArgOnPostConstruct         BikeErrorCode = 4
	InvalidNumArgOnDestroy               BikeErrorCode = 5
	ConstructorInvalidNumberReturnValues BikeErrorCode = 6
	ConstructorReturnNoPointerValue      BikeErrorCode = 7
	ComponentConstructorNull             BikeErrorCode = 8
)

type BikeError struct {
	messageError string
	errorCode    BikeErrorCode
}

func (_self *BikeError) Error() string {
	return _self.messageError
}

func (_self *BikeError) ErrorCode() BikeErrorCode {
	return _self.errorCode
}

type Bike struct {
	componentsByType map[reflect.Type]*Component
	componentsById   map[string]*Component
	components       []*Component
}

type Container struct {
	bike *Bike
}

func NewBike() *Bike {
	return &Bike{
		componentsByType: make(map[reflect.Type]*Component),
		componentsById:   make(map[string]*Component),
		components:       make([]*Component, 0),
	}
}

func (_self *Bike) Registry(component Component) error {

	// Check if component have not constructor method
	if component.Constructor == nil {
		return &BikeError{messageError: "Constructor must no be nill", errorCode: ComponentConstructorNull}
	}

	// Check Constructor component
	constructorType := reflect.TypeOf(component.Constructor)
	if constructorType.NumOut() != 1 {
		return &BikeError{messageError: "Constructor must return one value", errorCode: ConstructorInvalidNumberReturnValues}
	}
	typeComponent := constructorType.Out(0)

	// Check return Constructor value
	if typeComponent.Kind() != reflect.Pointer && typeComponent.Kind() != reflect.Interface {
		return &BikeError{messageError: "Constructor must return a pointer o interface value", errorCode: ConstructorReturnNoPointerValue}
	}

	// Check PostConstruct
	if len([]rune(component.PostConstruct)) > 0 {
		componentType := constructorType.Out(0)
		method, ok := componentType.MethodByName(component.PostConstruct)
		if !ok {
			return &BikeError{messageError: "Component.PostConstruct [" + component.PostConstruct + "] not found" + component.PostConstruct, errorCode: InvalidNumArgOnPostConstruct}
		}
		if method.Type.NumIn() != 1 {
			return &BikeError{messageError: "Invalid argument number of Component.PostConstruct. PostConstruct:" + component.PostConstruct, errorCode: InvalidNumArgOnPostConstruct}
		}
	}

	// Check Destroy
	if len([]rune(component.Destroy)) > 0 {
		componentType := constructorType.Out(0)
		method, ok := componentType.MethodByName(component.Destroy)
		if !ok {
			return &BikeError{messageError: "Invalid Component.Destroy:" + component.Destroy, errorCode: InvalidNumArgOnPostConstruct}
		}
		if method.Type.NumIn() != 1 {
			return &BikeError{messageError: "Invalid number arguments of Component.Destroy:" + component.Destroy, errorCode: InvalidNumArgOnPostConstruct}
		}
	}

	// Registry by id
	if len([]rune(component.Id)) > 0 {
		_self.componentsById[component.Id] = &component
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
		return &BikeError{messageError: message, errorCode: InvalidScope}
	}

	// Add to array
	_self.components = append(_self.components, &component)

	return nil
}

func (_self *Bike) instanceByTypeAny(inputType any) (interface{}, error) {
	_type := reflect.TypeOf(inputType)
	if _type.Kind() == reflect.Pointer && _type.Elem().Kind() == reflect.Interface {
		_type = _type.Elem()
	}
	return _self.instanceByType(_type)
}

func (_self *Bike) instanceById(id string) (interface{}, error) {
	component, ok := _self.componentsById[id]
	if ok {
		if component.Scope == Singleton {
			if component.instanceValue.Elem().CanAddr() {
				return component.instanceValue.Elem().Addr().Interface(), nil
			} else {
				return component.instanceValue.Elem().Interface(), nil
			}
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
	return nil, &BikeError{messageError: message, errorCode: ComponentNotFound}
}

func (_self *Bike) instanceByType(_type reflect.Type) (interface{}, error) {
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
	return nil, &BikeError{messageError: message, errorCode: ComponentNotFound}
}

func (_self *Bike) createComponent(component *Component) (*reflect.Value, error) {
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
			message := "Dependecy Type: [" + inputType.Name() + "] required by function" + constructorType.Name() + "] not found "
			return nil, &BikeError{messageError: message, errorCode: DependecyByTypeNotFound}
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

func (_self *Bike) Start() (*Container, error) {
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

func (_self *Bike) Stop() error {
	var lastError error
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

func (_self *Container) InstanceByType(inputType any) (interface{}, error) {
	return _self.bike.instanceByTypeAny(inputType)
}

func (_self *Container) InstanceById(id string) (interface{}, error) {
	return _self.bike.instanceById(id)
}
