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
	// ConstrutorReturnNotNilError return not nil error
	ConstrutorReturnNotNilError ErrorCode = 9
	// ConstructorLastReturnValueIsNotError error when last return isnt type error
	ConstructorLastReturnValueIsNotError ErrorCode = 10
	// PostStartWithScopeDifferentToSingleton error when a component has PostStart and Scope different to Singleton
	PostStartWithScopeDifferentToSingleton ErrorCode = 11
	// PostContructReturnError error when a PostContruct return a error
	PostContructReturnError = 12
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
	components []*Component
}

// Container struct with component management
type Container struct {
	componentsByType map[reflect.Type]*Component
	componentsByID   map[string]*Component
	components       []*Component
}

// NewBike create a Bike instance
func NewBike() *Bike {
	return &Bike{
		components: make([]*Component, 0),
	}
}

// Add add a component to bike
func (_self *Bike) Add(component Component) {
	// Add to array
	_self.components = append(_self.components, &component)
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

func validateComponent(component *Component) *Error {
	// Check if component have not constructor method
	if component.Constructor == nil {
		return &Error{
			messageError: fmt.Sprintf("Error on Component ID:[%s]. Constructor must not be nil", component.ID),
			errorCode:    ComponentConstructorNull}
	}

	// Check Scope
	if component.Scope != Singleton && component.Scope != Prototype {
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
		if method.Type.NumIn() != 1 {
			return &Error{
				messageError: fmt.Sprintf("Error on Component ID:[%s]. Invalid argument number of PostConstruct:[%s]", component.ID, component.PostConstruct),
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

func (_self *Container) instanceByTypeAny(inputType any) (interface{}, *Error) {
	_type := reflect.TypeOf(inputType)
	if _type.Kind() == reflect.Pointer && _type.Elem().Kind() == reflect.Interface {
		_type = _type.Elem()
	}
	return _self.instanceByType(_type)
}

func (_self *Container) instanceByID(id string) (interface{}, *Error) {
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

func (_self *Container) instanceByType(_type reflect.Type) (interface{}, *Error) {
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
		message = "Component by type:" + getTypeName(_type) + " not found"
	}
	return nil, &Error{messageError: message, errorCode: DependecyByTypeNotFound}
}

func (_self *Container) createComponent(component *Component) (*reflect.Value, *Error) {
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
			return nil, &Error{
				messageError: fmt.Sprintf("Error on Component ID:[%s]. Error to get dependecy: [%s] required by Constructor:[%s]", component.ID, getTypeName(inputType), getFuncName(component)),
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
				errorCode:    ConstrutorReturnNotNilError}
		}
	}

	component.instanceValue = &instanceResult[0]

	// Call PostConstruct
	componentType := constructorType.Out(0)
	if len([]rune(component.PostConstruct)) > 0 {
		method, _ := componentType.MethodByName(component.PostConstruct)
		typeError := reflect.TypeOf((*error)(nil)).Elem()
		returnValues := method.Func.Call([]reflect.Value{*component.instanceValue})
		for _, value := range returnValues {
			if value.Type().Implements(typeError) {
				err := value.Elem().Interface().(error)
				return nil, &Error{
					messageError: fmt.Sprintf("Error on Component ID:[%s]. PostConstruct return an error:[%s]", component.ID, err.Error()),
					errorCode:    PostContructReturnError}
			}
		}
	}

	return component.instanceValue, nil
}

// Start start bike
func (_self *Bike) Start() (*Container, *Error) {
	container := &Container{
		componentsByType: make(map[reflect.Type]*Component),
		componentsByID:   make(map[string]*Component),
		components:       _self.components,
	}

	for _, component := range container.components {
		// 1. Validate
		validateErr := validateComponent(component)
		if validateErr != nil {
			return nil, validateErr
		}

		// 2. Registry
		container.registry(component)

		// 3. Create component
		if component.Scope == Singleton {
			instanceValue, err := container.createComponent(component)
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
							errorCode:    PostContructReturnError}
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
								errorCode:    PostContructReturnError}
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
	return _self.instanceByTypeAny(inputType)
}

// InstanceByID return a instance by ID
func (_self *Container) InstanceByID(id string) (interface{}, *Error) {
	return _self.instanceByID(id)
}

func getTypeName(_type reflect.Type) string {
	if _type.Kind() == reflect.Pointer {
		return "*" + _type.Elem().Name()
	} else {
		return _type.Name()
	}
}

func getFuncName(component *Component) string {
	return runtime.FuncForPC(reflect.ValueOf(component.Constructor).Pointer()).Name()
}
