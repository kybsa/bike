// Package bike contains core features
package bike

import (
	"fmt"
	"reflect"
)

// Container struct with component management
type Container struct {
	componentsByType           map[reflect.Type]*Component
	componentsByID             map[string]*Component
	components                 []*Component
	customScopeInstancesByType map[Scope]map[string]map[reflect.Type]interface{}
	customScopeInstancesByID   map[Scope]map[string]map[string]interface{}
}

// Registry a component to Container
func (_self *Container) registry(component *Component) {
	// Registry by id
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
			mapContext := _self.customScopeInstancesByID[scope][idContext]
			if mapContext != nil {
				instanceByContext := mapContext[id]
				if instanceByContext != nil {
					return instanceByContext, nil
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
				_self.customScopeInstancesByType[scope][idContext] = make(map[reflect.Type]interface{})
			}
			constructorType := reflect.TypeOf(component.Constructor)
			typeComponent := constructorType.Out(0)
			_self.customScopeInstancesByType[scope][idContext][typeComponent] = interfaceInstance

			_, ok = _self.customScopeInstancesByID[scope][idContext]
			if !ok {
				_self.customScopeInstancesByID[scope][idContext] = make(map[string]interface{})
			}
			_self.customScopeInstancesByID[scope][idContext][component.ID] = interfaceInstance
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
					return instanceByContext, nil
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
			// TODO El problema es que "component" almacenado en customScope no tiene definido instanceValue
			//  Entonces siempre se regresa nulo, sin embargo
			// PUEDE QUe lo mejor sea almacenar en no el Component sino el value o
			// de lo contrario crear una copia del componente!!!!
			if !ok {
				_self.customScopeInstancesByType[scope][idContext] = make(map[reflect.Type]interface{})
			}
			_self.customScopeInstancesByType[scope][idContext][_type] = interfaceInstance

			_, ok = _self.customScopeInstancesByID[scope][idContext]
			if !ok {
				_self.customScopeInstancesByID[scope][idContext] = make(map[string]interface{})
			}
			_self.customScopeInstancesByID[scope][idContext][component.ID] = interfaceInstance
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

// InstanceByTypeAndIDContext return a instance by type, scope and idContext
func (_self *Container) InstanceByTypeAndIDContext(inputType any, scope Scope, idContext string) (interface{}, *Error) {
	return _self.instanceByTypeAny(inputType, scope, idContext)
}

// InstanceByIDAndIDContext return a instance by id, scope and idContext
func (_self *Container) InstanceByIDAndIDContext(id string, scope Scope, idContext string) (interface{}, *Error) {
	return _self.instanceByID(id, scope, idContext)
}

func (_self *Container) RemoveContext(scope Scope, idContext string) *Error {
	// Remove context or error if doesn't exist
	if _, ok := _self.customScopeInstancesByID[scope]; !ok {
		return &Error{
			messageError: fmt.Sprintf("Invalid Scope:[%d]", scope),
			errorCode:    InvalidScope}
	}
	if _, ok := _self.customScopeInstancesByID[scope][idContext]; !ok {
		return &Error{
			messageError: fmt.Sprintf("Context id:[%s] not found", idContext),
			errorCode:    InvalidScope}
	}
	delete(_self.customScopeInstancesByID[scope], idContext)
	return nil
}
