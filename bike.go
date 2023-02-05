// Package bike contains core features
package bike

import (
	"fmt"
	"reflect"
	"runtime"
	"sync"

	"github.com/google/uuid"
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
	PostStart               string
	instanceValue           *reflect.Value
	prototypeInstancesValue []*reflect.Value
}

// ErrorCode type to enum error codes
type ErrorCode uint8

const (
	// DependencyByIDNotFound error code when a dependency not found by id
	DependencyByIDNotFound ErrorCode = 0
	// DependencyByTypeNotFound error code when a dependency not found by type
	DependencyByTypeNotFound ErrorCode = 1
	// InvalidScope error code when a invalid scope
	InvalidScope ErrorCode = 2
	// InvalidNumArgOnPostConstruct error code when PostConstruct method has arguments
	InvalidNumArgOnPostConstruct ErrorCode = 4
	// InvalidNumArgOnDestroy error code when Destroy method has arguments
	InvalidNumArgOnDestroy ErrorCode = 5
	// InvalidNumberOfReturnValuesOnConstructor error when constructor return invalid number of values
	InvalidNumberOfReturnValuesOnConstructor ErrorCode = 6
	// ConstructorReturnNoPointerValue error when constructor method return a no pointer value
	ConstructorReturnNoPointerValue ErrorCode = 7
	// ComponentConstructorNull error when Constructor property is null
	ComponentConstructorNull ErrorCode = 8
	// ConstructorReturnNotNilError return not nil error
	ConstructorReturnNotNilError ErrorCode = 9
	// ConstructorLastReturnValueIsNotError error when last return isn't type error
	ConstructorLastReturnValueIsNotError ErrorCode = 10
	// PostStartWithScopeDifferentToSingleton error when a component has PostStart and Scope different to Singleton
	PostStartWithScopeDifferentToSingleton ErrorCode = 11
	// PostConstructReturnError error when a PostConstruct return a error
	PostConstructReturnError ErrorCode = 12
	// DuplicateScope error when a scope exist
	DuplicateScope ErrorCode = 13
)

// Error struct with error info
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
	components   []*Component
	customScopes map[Scope]string
}

// Container struct with component management
type Container struct {
	componentsByType map[reflect.Type]*Component
	componentsByID   map[string]*Component
	components       []*Component

	customScopeInstancesByType map[Scope]map[string]map[reflect.Type]*Component // map[IdScope][IdContext][reflect.Type]
	customScopeInstancesById   map[Scope]map[string]map[string]*Component       // map[IdScope][IdContext][reflect.Type]

}

// NewBike create a Bike instance
func NewBike() *Bike {
	return &Bike{
		components:   make([]*Component, 0),
		customScopes: make(map[Scope]string, 0),
	}
}

// Add add a component to bike
func (_self *Bike) Add(component Component) {
	// Add to array
	_self.components = append(_self.components, &component)
}

func (_self *Bike) AddCustomScope(newScope Scope, name string) *Error {
	if newScope == Singleton || newScope == Prototype {
		return &Error{
			errorCode:    DuplicateScope,
			messageError: fmt.Sprintf("Duplicated scope id:%d name:[%s]", newScope, name),
		}
	}
	if _, ok := _self.customScopes[newScope]; ok {
		return &Error{
			errorCode:    DuplicateScope,
			messageError: fmt.Sprintf("Duplicated scope id:%d name:[%s]", newScope, name),
		}
	}
	_self.customScopes[newScope] = name
	return nil
}

// Registry a component to Container
func (_self *Container) registry(component *Component) {
	// Registry by id
	if len([]rune(component.ID)) == 0 {
		component.ID = uuid.NewString()
	}
	_self.componentsByID[component.ID] = component

	constructorType := reflect.TypeOf(component.Constructor)
	typeComponent := constructorType.Out(0)
	_self.componentsByType[typeComponent] = component

	// Registry by interfaces
	for _, inter := range component.Interfaces {
		interfaceType := reflect.TypeOf(inter).Elem()
		_self.componentsByType[interfaceType] = component
	}

	// Init array of prototype instances
	if component.Scope == Prototype {
		component.prototypeInstancesValue = make([]*reflect.Value, 0)
	}
}

func (_self *Bike) validateComponent(component *Component) *Error {
	// Check if component have not constructor method
	if component.Constructor == nil {
		return &Error{
			messageError: fmt.Sprintf("Error on Component ID:[%s]. Constructor must not be nil", component.ID),
			errorCode:    ComponentConstructorNull}
	}

	// Check Scope
	_, isCustomScope := _self.customScopes[component.Scope]
	if component.Scope != Singleton && component.Scope != Prototype && !isCustomScope {
		return &Error{
			messageError: fmt.Sprintf("Error on Component ID:[%s]. Invalid Scope: %s", component.ID, component.Scope.String()),
			errorCode:    InvalidScope}
	}

	// Check Constructor component
	constructorType := reflect.TypeOf(component.Constructor)
	if constructorType.NumOut() == 2 {
		typeErrorReturn := constructorType.Out(1)
		errorType := reflect.TypeOf((*error)(nil)).Elem()
		if !typeErrorReturn.Implements(errorType) {
			return &Error{
				messageError: fmt.Sprintf("Error on Component ID:[%s]. Last return value must be of error type", component.ID),
				errorCode:    ConstructorLastReturnValueIsNotError}
		}
	} else if constructorType.NumOut() != 1 {
		return &Error{
			messageError: fmt.Sprintf("Error on Component ID:[%s]. Constructor must return one value", component.ID),
			errorCode:    InvalidNumberOfReturnValuesOnConstructor}
	}
	typeComponent := constructorType.Out(0)

	// Check return Constructor value
	if typeComponent.Kind() != reflect.Pointer && typeComponent.Kind() != reflect.Interface {
		return &Error{
			messageError: fmt.Sprintf("Error on Component ID:[%s]. Constructor must return a pointer o interface value", component.ID),
			errorCode:    ConstructorReturnNoPointerValue}
	}

	// Check PostConstruct
	if len([]rune(component.PostConstruct)) > 0 {
		componentType := constructorType.Out(0)
		method, ok := componentType.MethodByName(component.PostConstruct)
		if !ok {
			return &Error{
				messageError: fmt.Sprintf("Error on Component ID:[%s]. PostConstruct [%s] not found", component.ID, component.PostConstruct),
				errorCode:    InvalidNumArgOnPostConstruct}
		}
		methodType := method.Type
		if methodType.NumIn() == 2 {
			inputType := methodType.In(1)
			if inputType != reflect.TypeOf((*Container)(nil)) {
				return &Error{
					messageError: fmt.Sprintf("Error on Component ID:[%s]. Invalid argument type of PostConstruct:[%s], expected *Container, actual:%s", component.ID, component.PostConstruct, getTypeName(inputType)),
					errorCode:    InvalidNumArgOnPostConstruct}
			}
		} else if method.Type.NumIn() != 1 {
			return &Error{
				messageError: fmt.Sprintf("Error on Component ID:[%s]. Invalid argument number of PostConstruct:[%s], expected 0 or 1 arguments, actual:%d", component.ID, component.PostConstruct, method.Type.NumIn()),
				errorCode:    InvalidNumArgOnPostConstruct}
		}
	}

	// Check Destroy
	if len([]rune(component.Destroy)) > 0 {
		componentType := constructorType.Out(0)
		method, ok := componentType.MethodByName(component.Destroy)
		if !ok {
			return &Error{
				messageError: fmt.Sprintf("Error on Component ID:[%s]. Invalid Component.Destroy:%s", component.ID, component.Destroy),
				errorCode:    InvalidNumArgOnPostConstruct}
		}
		if method.Type.NumIn() != 1 {
			return &Error{
				messageError: fmt.Sprintf("Error on Component ID:[%s]. Invalid number arguments of Destroy:[%s]", component.ID, component.Destroy),
				errorCode:    InvalidNumArgOnPostConstruct}
		}
	}

	// Check PostStart
	if len([]rune(component.PostStart)) > 0 {

		if component.Scope != Singleton {
			return &Error{
				messageError: fmt.Sprintf("Error on Component ID:[%s]. PostStart [%s] only supported when scope equal to Singleton", component.ID, component.PostStart),
				errorCode:    PostStartWithScopeDifferentToSingleton}
		}

		componentType := constructorType.Out(0)
		method, ok := componentType.MethodByName(component.PostStart)
		if !ok {
			return &Error{
				messageError: fmt.Sprintf("Error on Component ID:[%s]. PostStart [%s] not found", component.ID, component.PostStart),
				errorCode:    InvalidNumArgOnPostConstruct}
		}
		if method.Type.NumIn() != 1 {
			return &Error{
				messageError: fmt.Sprintf("Error on Component ID:[%s]. Invalid argument number of PostStart [%s]", component.ID, component.PostStart),
				errorCode:    InvalidNumArgOnPostConstruct}
		}
	}

	return nil
}

func (_self *Container) instanceByTypeAny(inputType any, scope Scope, idContext string) (interface{}, *Error) {
	_type := reflect.TypeOf(inputType)
	if _type.Kind() == reflect.Pointer && _type.Elem().Kind() == reflect.Interface {
		_type = _type.Elem()
	}
	return _self.instanceByType(_type, scope, idContext)
}

func (_self *Container) instanceByID(id string, scope Scope, idContext string) (interface{}, *Error) {
	component, ok := _self.componentsByID[id]
	if ok {
		if component.Scope == Singleton {
			if component.instanceValue.Elem().CanAddr() {
				return component.instanceValue.Elem().Addr().Interface(), nil
			}
			return component.instanceValue.Elem().Interface(), nil
		}

		if component.Scope != Prototype {
			mapContext := _self.customScopeInstancesById[scope][idContext]
			if mapContext != nil {
				instanceByContext := mapContext[id]
				if instanceByContext != nil {
					return instanceByContext.instanceValue, nil
				}
			}
		}

		instance, err := _self.createComponent(component, scope, idContext)
		if err != nil {
			return nil, err
		}
		var interfaceInstance any
		if instance.Elem().CanAddr() {
			interfaceInstance = instance.Elem().Addr().Interface()
		} else {
			interfaceInstance = instance.Elem().Interface()
		}

		if component.Scope == Prototype {
			component.prototypeInstancesValue = append(component.prototypeInstancesValue, instance)
			return interfaceInstance, nil
		} else {
			_, ok := _self.customScopeInstancesByType[scope][idContext]
			if !ok {
				_self.customScopeInstancesByType[scope][idContext] = make(map[reflect.Type]*Component)
			}
			constructorType := reflect.TypeOf(component.Constructor)
			typeComponent := constructorType.Out(0)
			_self.customScopeInstancesByType[scope][idContext][typeComponent] = component

			_, ok = _self.customScopeInstancesById[scope][idContext]
			if !ok {
				_self.customScopeInstancesById[scope][idContext] = make(map[string]*Component)
			}
			_self.customScopeInstancesById[scope][idContext][component.ID] = component
		}
		return interfaceInstance, nil
	}
	message := "Component by id:" + id + " not found"
	return nil, &Error{messageError: message, errorCode: DependencyByIDNotFound}
}

func (_self *Container) instanceByType(_type reflect.Type, scope Scope, idContext string) (interface{}, *Error) {
	component, ok := _self.componentsByType[_type]
	if ok {
		if component.Scope == Singleton {
			return component.instanceValue.Elem().Addr().Interface(), nil
		}

		if component.Scope != Prototype {
			mapContext := _self.customScopeInstancesByType[scope][idContext]
			if mapContext != nil {
				instanceByContext := mapContext[_type]
				if instanceByContext != nil {
					return instanceByContext.instanceValue, nil
				}
			}
		}

		instance, err := _self.createComponent(component, scope, idContext)
		if err != nil {
			return nil, err
		}
		var interfaceInstance any
		if instance.Elem().CanAddr() {
			interfaceInstance = instance.Elem().Addr().Interface()
		} else {
			interfaceInstance = instance.Elem().Interface()
		}
		if component.Scope == Prototype {
			component.prototypeInstancesValue = append(component.prototypeInstancesValue, instance)
		} else {
			_, ok := _self.customScopeInstancesByType[scope][idContext]
			if !ok {
				_self.customScopeInstancesByType[scope][idContext] = make(map[reflect.Type]*Component)
			}
			_self.customScopeInstancesByType[scope][idContext][_type] = component

			_, ok = _self.customScopeInstancesById[scope][idContext]
			if !ok {
				_self.customScopeInstancesById[scope][idContext] = make(map[string]*Component)
			}
			_self.customScopeInstancesById[scope][idContext][component.ID] = component
		}
		return interfaceInstance, nil
	}
	var message string
	if _type.Kind() == reflect.Interface {
		message = "Component by type:" + _type.Name() + " not found"

	} else {
		message = "Component by type:" + getTypeName(_type) + " not found"
	}
	return nil, &Error{messageError: message, errorCode: DependencyByTypeNotFound}
}

func (_self *Container) createComponent(component *Component, scope Scope, idContext string) (*reflect.Value, *Error) {
	// Create component by constructor method
	constructorValue := reflect.ValueOf(component.Constructor)
	constructorType := reflect.TypeOf(component.Constructor)

	// Search dependencies
	args := make([]reflect.Value, constructorType.NumIn())
	for i := 0; i < constructorType.NumIn(); i++ {
		inputType := constructorType.In(i)
		inputArg, err := _self.instanceByType(inputType, scope, idContext)
		if err == nil {
			args[i] = reflect.ValueOf(inputArg)
		} else {
			return nil, &Error{
				messageError: fmt.Sprintf("Error on Component ID:[%s]. Error to get dependency: [%s] required by Constructor:[%s]", component.ID, getTypeName(inputType), getFuncName(component)),
				errorCode:    err.ErrorCode()}
		}
	}
	// Create component with dependencies
	instanceResult := constructorValue.Call(args)

	if len(instanceResult) == 2 {
		errorElem := instanceResult[1].Elem()
		if !instanceResult[1].IsNil() && errorElem.Interface() != nil {
			constructorError := errorElem.Interface().(error)
			return nil, &Error{
				messageError: fmt.Sprintf("Error on Component ID:[%s]. Constructor return an error:[%s]", component.ID, constructorError.Error()),
				errorCode:    ConstructorReturnNotNilError}
		}
	}

	instanceValue := &instanceResult[0]

	// Call PostConstruct
	componentType := constructorType.Out(0)
	if len([]rune(component.PostConstruct)) > 0 {
		method, _ := componentType.MethodByName(component.PostConstruct)
		typeError := reflect.TypeOf((*error)(nil)).Elem()

		var in []reflect.Value
		if method.Type.NumIn() == 2 {
			in = []reflect.Value{*instanceValue, reflect.ValueOf(_self)}
		} else {
			in = []reflect.Value{*instanceValue}
		}

		returnValues := method.Func.Call(in)
		for _, value := range returnValues {
			if value.Type().Implements(typeError) {
				err := value.Elem().Interface().(error)
				return nil, &Error{
					messageError: fmt.Sprintf("Error on Component ID:[%s]. PostConstruct return an error:[%s]", component.ID, err.Error()),
					errorCode:    PostConstructReturnError}
			}
		}
	}

	return instanceValue, nil
}

// Start start bike
func (_self *Bike) Start() (*Container, *Error) {
	container := &Container{
		componentsByType:           make(map[reflect.Type]*Component),
		componentsByID:             make(map[string]*Component),
		components:                 _self.components,
		customScopeInstancesByType: make(map[Scope]map[string]map[reflect.Type]*Component),
		customScopeInstancesById:   make(map[Scope]map[string]map[string]*Component),
	}

	// 0. Create map with custom scopes
	for key, _ := range _self.customScopes {
		container.customScopeInstancesByType[key] = make(map[string]map[reflect.Type]*Component)
		container.customScopeInstancesById[key] = make(map[string]map[string]*Component)
	}

	for _, component := range container.components {
		// 1. Validate
		validateErr := _self.validateComponent(component)
		if validateErr != nil {
			return nil, validateErr
		}

		// 2. Registry
		container.registry(component)

		// 3. Create component
		if component.Scope == Singleton {
			instanceValue, err := container.createComponent(component, Singleton, "0")
			if err != nil {
				return nil, err
			}
			component.instanceValue = instanceValue
		}
	}

	// 4. PostStart
	var wg sync.WaitGroup
	for _, component := range container.components {
		if component.Scope == Singleton && len([]rune(component.PostStart)) > 0 {
			constructorType := reflect.TypeOf(component.Constructor)
			componentType := constructorType.Out(0)
			method, _ := componentType.MethodByName(component.PostStart)
			wg.Add(1)
			go func(internalComponent *Component) {
				defer wg.Done()
				method.Func.Call([]reflect.Value{*internalComponent.instanceValue})
			}(component)
		}
	}
	wg.Wait()

	return container, nil
}

// Stop stop container
func (_self *Container) Stop() *Error {
	typeError := reflect.TypeOf((*error)(nil)).Elem()
	for _, component := range _self.components {
		if len([]rune(component.Destroy)) > 0 {
			componentType := reflect.TypeOf(component.Constructor).Out(0)
			method, _ := componentType.MethodByName(component.Destroy)
			if component.Scope == Singleton {
				returnValues := method.Func.Call([]reflect.Value{*component.instanceValue})
				for _, value := range returnValues {
					if value.Type().Implements(typeError) {
						err := value.Elem().Interface().(error)
						return &Error{
							messageError: fmt.Sprintf("Error on Component ID:[%s]. Destroy return an error:[%s]", component.ID, err.Error()),
							errorCode:    PostConstructReturnError}
					}
				}
			} else if component.Scope == Prototype {
				for _, prototypeInstance := range component.prototypeInstancesValue {
					returnValues := method.Func.Call([]reflect.Value{*prototypeInstance})
					for _, value := range returnValues {
						if value.Type().Implements(typeError) {
							err := value.Elem().Interface().(error)
							return &Error{
								messageError: fmt.Sprintf("Error on Component ID:[%s]. Destroy return an error:[%s]", component.ID, err.Error()),
								errorCode:    PostConstructReturnError}
						}
					}
				}
			}
		}
	}
	return nil
}

// InstanceByType return a instance by type
func (_self *Container) InstanceByType(inputType any) (interface{}, *Error) {
	return _self.instanceByTypeAny(inputType, Singleton, "0")
}

// InstanceByID return a instance by ID
func (_self *Container) InstanceByID(id string) (interface{}, *Error) {
	return _self.instanceByID(id, Singleton, "0")
}

func getTypeName(_type reflect.Type) string {
	if _type.Kind() == reflect.Pointer {
		if _type.Elem().Kind() == reflect.Pointer {
			return "*" + getTypeName(_type.Elem())
		}
		return "*" + _type.Elem().Name()
	} else {
		return _type.Name()
	}
}

func getFuncName(component *Component) string {
	return runtime.FuncForPC(reflect.ValueOf(component.Constructor).Pointer()).Name()
}

// InstanceByType return a instance by type
func (_self *Container) InstanceByTypeAndIdContext(inputType any, scope Scope, idContext string) (interface{}, *Error) {
	return _self.instanceByTypeAny(inputType, scope, idContext)
}

// InstanceById return a instance by type
func (_self *Container) InstanceByIdAndIdContext(id string, scope Scope, idContext string) (interface{}, *Error) {
	return _self.instanceByID(id, scope, idContext)
}
