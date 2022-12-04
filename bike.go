package bike

import (
	"fmt"
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
	Type                    any
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
	NotFound                             BikeErrorCode = 0
	DependecyByIdNotFound                BikeErrorCode = 1
	NullDependencyConfigType             BikeErrorCode = 2
	DependecyByTypeNotFound              BikeErrorCode = 3
	InvalidField                         BikeErrorCode = 4
	ComponentTypeNull                    BikeErrorCode = 5
	InvalidScope                         BikeErrorCode = 6
	ComponentNotFound                    BikeErrorCode = 7
	InvalidNumArgOnPostConstruct         BikeErrorCode = 8
	InvalidNumArgOnDestroy               BikeErrorCode = 9
	ConstructorInvalidNumberReturnValues BikeErrorCode = 10
	ConstructorReturnNoPointerValue      BikeErrorCode = 11
)

type BikeError struct {
	messageError string
	errorCode    BikeErrorCode
}

func (_self *BikeError) Error() string {
	return _self.messageError
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

	// Registry by id
	if len([]rune(component.Id)) > 0 {
		_self.componentsById[component.Id] = &component
	}

	// Registry by Type
	if component.Type != nil {
		componentType := reflect.TypeOf(component.Type)
		_self.componentsByType[componentType] = &component
	}

	// If component have constructor method
	if component.Constructor != nil {
		constructorType := reflect.TypeOf(component.Constructor)
		if constructorType.NumOut() != 1 {
			return &BikeError{messageError: "Constructor must return one value", errorCode: ConstructorInvalidNumberReturnValues}
		}
		typeComponent := constructorType.Out(0)

		if typeComponent.Kind() != reflect.Pointer && typeComponent.Kind() != reflect.Interface {
			return &BikeError{messageError: "Constructor must return a pointer o interface value", errorCode: ConstructorReturnNoPointerValue}
		}

		_self.componentsByType[typeComponent] = &component
	}

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
	fmt.Println(component)
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
	if component.Constructor != nil {
		constructorValue := reflect.ValueOf(component.Constructor)
		constructorType := reflect.TypeOf(component.Constructor)

		if constructorType.NumOut() != 1 {
			return nil, &BikeError{messageError: "Constructor must return one value", errorCode: ConstructorInvalidNumberReturnValues}
		}

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
		return component.instanceValue, nil
	}

	if component.Type == nil {
		return nil, &BikeError{messageError: "Component.Type must be not null", errorCode: ComponentTypeNull}
	}

	componentType := reflect.TypeOf(component.Type)
	instanceValue := reflect.New(componentType.Elem())

	//  Inject dependencies using field Component.Dependencies
	for fieldName, dependencyConfig := range component.Dependencies {
		dependencyType := reflect.TypeOf(dependencyConfig.Type)
		err := _self.injectDependency(&instanceValue, &componentType, fieldName, dependencyConfig.Id, &dependencyType)
		if err != nil {
			return nil, err
		}
	}

	// Inject dependecies using struct tags
	for i := 0; i < componentType.Elem().NumField(); i++ {
		field := componentType.Elem().Field(i)
		if _, ok := field.Tag.Lookup("inject"); ok {
			fieldName := field.Name
			id, _ := field.Tag.Lookup("id")
			typeField := field.Type
			err := _self.injectDependency(&instanceValue, &componentType, fieldName, id, &typeField)
			if err != nil {
				return nil, err
			}
		}
	}

	// Call init methods
	// Search Components methods to pointer struct
	if len([]rune(component.PostConstruct)) > 0 {
		method, ok := componentType.MethodByName(component.PostConstruct)
		if !ok {
			return nil, &BikeError{messageError: "Component.PostConstruct [" + component.PostConstruct + "] not found" + component.PostConstruct, errorCode: InvalidNumArgOnPostConstruct}
		}
		if method.Type.NumIn() != 1 {
			return nil, &BikeError{messageError: "Invalid argument number of Component.PostConstruct. PostConstruct:" + component.PostConstruct, errorCode: InvalidNumArgOnPostConstruct}
		}
		method.Func.Call([]reflect.Value{instanceValue})
	}

	return &instanceValue, nil
}

func (_self *Bike) injectDependency(instanceValue *reflect.Value, componentType *reflect.Type, key string, id string, dependencyType *reflect.Type) error {
	var dependency *reflect.Value = nil
	// Inject by id dependency
	if len([]rune(id)) > 0 {
		dependencyById, okById := _self.componentsById[id]
		if okById {
			if dependencyById.Scope == Singleton {
				dependency = dependencyById.instanceValue
			} else if dependencyById.Scope == Prototype {
				var err error
				dependency, err = _self.createComponent(dependencyById)
				dependencyById.prototypeInstancesValue = append(dependencyById.prototypeInstancesValue, dependency)
				if err != nil {
					return err
				}
			} else {
				message := "Component with id " + id + " using invalid Scope: " + dependencyById.Scope.String()
				return &BikeError{messageError: message, errorCode: InvalidScope}
			}
		} else {
			message := "Dependecy: [" + key + "] required by [" + (*componentType).Elem().Name() + "] with id [" + id + "] not found "
			return &BikeError{messageError: message, errorCode: DependecyByIdNotFound}
		}
	}

	// Inject by Type
	if dependencyType == nil {
		message := "Invalid dependecy with Type nil"
		return &BikeError{messageError: message, errorCode: NullDependencyConfigType}
	} else if dependency == nil {
		dependencyByType, okByType := _self.componentsByType[*dependencyType]
		if okByType {
			if dependencyByType.Scope == Singleton {
				dependency = dependencyByType.instanceValue
			} else if dependencyByType.Scope == Prototype {
				var err error
				dependency, err = _self.createComponent(dependencyByType)
				dependencyByType.prototypeInstancesValue = append(dependencyByType.prototypeInstancesValue, dependency)
				if err != nil {
					return err
				}
			} else {
				message := "Component with type " + (*componentType).Elem().Name() + " using invalid Scope: " + dependencyByType.Scope.String()
				return &BikeError{messageError: message, errorCode: InvalidScope}
			}
		}
	}

	if dependency == nil {
		message := "Dependecy: [" + key + "] required by [" + (*componentType).Elem().Name() + "] not found "
		return &BikeError{messageError: message, errorCode: DependecyByTypeNotFound}
	}

	fieldValue := instanceValue.Elem().FieldByName(key)
	if !fieldValue.IsValid() {
		message := "Field: [" + key + "] no found in [" + (*componentType).Elem().Name() + "] "
		return &BikeError{messageError: message, errorCode: InvalidField}
	}
	fieldValue.Set(dependency.Elem().Addr())

	return nil
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
			componentType := reflect.TypeOf(component.Type)
			method, ok := componentType.MethodByName(component.Destroy)
			if !ok {
				lastError = &BikeError{messageError: "Error to stop bike. Invalid Component.Destroy:" + component.Destroy, errorCode: InvalidNumArgOnPostConstruct}
			}
			if method.Type.NumIn() != 1 {
				lastError = &BikeError{messageError: "Error to stop bike. Invalid number arguments of Component.Destroy:" + component.Destroy, errorCode: InvalidNumArgOnPostConstruct}
				continue
			}
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
