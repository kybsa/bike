// Package bike contains core features
package bike

import (
	"fmt"
	"reflect"
	"runtime"
	"sync"
)

// Bike is main struct of this package
type Bike struct {
	components   []*Component
	customScopes map[Scope]string
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

// Start start bike
func (_self *Bike) Start() (*Container, *Error) {
	container := &Container{
		componentsByType:           make(map[reflect.Type]*Component),
		componentsByID:             make(map[string]*Component),
		components:                 _self.components,
		customScopeInstancesByType: make(map[Scope]map[string]map[reflect.Type]*Component),
		customScopeInstancesByID:   make(map[Scope]map[string]map[string]*Component),
	}

	// 0. Create map with custom scopes
	for key := range _self.customScopes {
		container.customScopeInstancesByType[key] = make(map[string]map[reflect.Type]*Component)
		container.customScopeInstancesByID[key] = make(map[string]map[string]*Component)
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
