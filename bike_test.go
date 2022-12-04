package bike

import (
	"testing"
)

type NoImplemented interface {
	NoImplemented() error
}

type InterfaceComponent interface {
	DoAnything()
}

type InterfaceAnyComponent interface {
	DoAny()
}

type StrucComponent struct {
	InitStatus bool
	StopStatus bool
}

func (_self *StrucComponent) DoAnything() {

}

func (_self *StrucComponent) Init() {
	_self.InitStatus = true
}

func (_self *StrucComponent) Stop() {
	_self.StopStatus = true
}

func (_self *StrucComponent) DoAny() {
}

func NewInterfaceComponent() InterfaceComponent {
	return &StrucComponent{}
}

func NewValueComponent() StrucComponent {
	return StrucComponent{}
}

func InvalidConstructor() {
}

func TestRegistry_GivenComponentWithId_WhenRegistry_ThenReturnNotNullInstanceById(t *testing.T) {
	// Given
	structComponent := Component{
		Type: (*StrucComponent)(nil),
		Id:   "IdStructComponent",
	}
	bike := NewBike()
	// When
	bike.Registry(structComponent)
	container, _ := bike.Start()
	// Then
	instance, err := container.InstanceById(structComponent.Id)
	if err != nil {
		t.Errorf("Error to get instance by id:%s.", err.Error())
	}
	if instance == nil {
		t.Errorf("InstanceById return nil.")
	}
}

func TestRegistry_GivenComponentWithType_WhenRegistry_ThenReturnNotNullInstanceByType(t *testing.T) {
	// Given
	structComponent := Component{
		Type: (*StrucComponent)(nil),
	}
	bike := NewBike()
	// When
	bike.Registry(structComponent)
	container, _ := bike.Start()
	// Then
	instance, err := container.InstanceByType((*StrucComponent)(nil))
	if err != nil {
		t.Errorf("Error to get instance by type:%s.", err.Error())
	}
	if instance == nil {
		t.Errorf("InstanceByType return nil.")
	}
}

func TestRegistry_GivenComponentWithTypeAndInterfaces_WhenRegistry_ThenReturnNotNullInstanceByInterfaceType(t *testing.T) {
	// Given
	structComponent := Component{
		Type:       (*StrucComponent)(nil),
		Interfaces: []any{(*InterfaceComponent)(nil)},
	}
	bike := NewBike()
	// When
	bike.Registry(structComponent)
	container, _ := bike.Start()
	// Then
	instance, err := container.InstanceByType((*InterfaceComponent)(nil))
	if err != nil {
		t.Errorf("Error to get instance by type:%s.", err.Error())
	}
	if instance == nil {
		t.Errorf("InstanceByType return nil.")
	}
}

func TestStart_GivenComponentWithScopePrototype_WhenStart_ThenCallInitMethod(t *testing.T) {
	// Given
	structComponent := Component{
		Type:          (*StrucComponent)(nil),
		Scope:         Prototype,
		PostConstruct: "Init",
	}
	bike := NewBike()
	// When
	bike.Registry(structComponent)
	container, _ := bike.Start()
	instance, _ := container.InstanceByType((*StrucComponent)(nil))
	strucComponentInstance := (instance).(*StrucComponent)
	bike.Stop()

	// Then
	if strucComponentInstance.InitStatus == false {
		t.Errorf("Bike doesn't call init method StrucComponent")
	}
}

func TestStop_GivenComponentWithScopePrototype_WhenStop_ThenCallDestroyMethod(t *testing.T) {
	// Given
	structComponent := Component{
		Type:    (*StrucComponent)(nil),
		Scope:   Prototype,
		Destroy: "Stop",
	}
	bike := NewBike()
	// When
	bike.Registry(structComponent)
	container, _ := bike.Start()
	instance, _ := container.InstanceByType((*StrucComponent)(nil))
	strucComponentInstance := (instance).(*StrucComponent)
	strucComponentInstance.StopStatus = false
	bike.Stop()
	// Then
	if strucComponentInstance.StopStatus == false {
		t.Errorf("Bike doesn't call stop method StrucComponent")
	}
}

func TestInstanceById_GivenComponentWithConstructorAndIdAndScopeSingleton_WhenRegistry_ThenReturnNotNullInstanceById(t *testing.T) {
	// Given
	interfaceComponent := Component{
		Constructor: NewInterfaceComponent,
		Scope:       Singleton,
		Id:          "IdComponent",
	}
	bike := NewBike()
	// When
	bike.Registry(interfaceComponent)
	container, _ := bike.Start()
	// Then
	instance, err := container.InstanceById(interfaceComponent.Id)
	if err != nil {
		t.Errorf("Error to get instance by id:%s.", err.Error())
	}
	if instance == nil {
		t.Errorf("InstanceById return nil.")
	}
}

func TestInstanceById_GivenComponentWithConstructorAndIdAndScopePrototype_WhenInstanceById_ThenReturnNotNull(t *testing.T) {
	// Given
	interfaceComponent := Component{
		Constructor: NewInterfaceComponent,
		Scope:       Prototype,
		Id:          "IdComponent",
	}
	bike := NewBike()
	bike.Registry(interfaceComponent)
	container, _ := bike.Start()
	// When
	instance, err := container.InstanceById(interfaceComponent.Id)
	// Then
	if err != nil {
		t.Errorf("Error to get instance by id:%s.", err.Error())
	}
	if instance == nil {
		t.Errorf("InstanceById return nil.")
	}
}

func TestInstanceById_GivenComponentWithIdAndScopePrototype_WhenInstanceById_ThenReturnNotNull(t *testing.T) {
	// Given
	structComponent := Component{
		Type:  (*StrucComponent)(nil),
		Scope: Prototype,
		Id:    "IdComponent",
	}
	bike := NewBike()
	bike.Registry(structComponent)
	container, _ := bike.Start()
	// When
	instance, err := container.InstanceById(structComponent.Id)
	// Then
	if err != nil {
		t.Errorf("Error to get instance by id:%s.", err.Error())
	}
	if instance == nil {
		t.Errorf("InstanceById return nil.")
	}
}

func TestInstanceById_GivenComponentWithInvalidConstructor_WhenInstanceById_ThenReturnError(t *testing.T) {
	// Given
	structComponent := Component{
		Constructor: InvalidConstructor,
		Scope:       Prototype,
		Id:          "IdComponent",
	}
	bike := NewBike()
	bike.Registry(structComponent)
	container, _ := bike.Start()
	// When
	_, err := container.InstanceById(structComponent.Id)
	// Then
	if err == nil {
		t.Errorf("InstanceById must return an error ")
	}
}

func TestInstanceById_GivenComponentWithInvalidScope_WhenInstanceById_ThenReturnError(t *testing.T) {
	// Given
	structComponent := Component{
		Constructor: InvalidConstructor,
		Scope:       99,
		Id:          "IdComponent",
	}
	bike := NewBike()
	bike.Registry(structComponent)
	container, _ := bike.Start()
	// When
	_, err := container.InstanceById(structComponent.Id)
	// Then
	if err == nil {
		t.Errorf("InstanceById must return an error ")
	}
}

func TestInstanceById_GivenComponentWithInvalidId_WhenInstanceById_ThenReturnError(t *testing.T) {
	// Given
	bike := NewBike()
	container, _ := bike.Start()
	// When
	_, err := container.InstanceById("any-id")
	// Then
	if err == nil {
		t.Errorf("InstanceById must return an error ")
	}
}

func TestInstancebyType_GivenComponentImplementInterface_WhenInstanceByType_ThenReturnInstance(t *testing.T) {
	// Given
	interfaceComponent := Component{
		Constructor: NewInterfaceComponent,
		Scope:       Prototype,
		Id:          "IdComponent",
	}
	bike := NewBike()
	bike.Registry(interfaceComponent)
	container, _ := bike.Start()
	// When
	instance, err := container.InstanceByType((*InterfaceComponent)(nil))
	// Then
	if err != nil {
		t.Errorf("InstanceByType must no return an error ")
	}
	if instance == nil {
		t.Errorf("InstanceByType must return not nil value ")
	}
}

func TestInstanceByType_GivenComponentScopePrototypeWithInvalidConstructorWhenInstanceByTypeThenReturError(t *testing.T) {
	// Given
	interfaceComponent := Component{
		Constructor: func(strucComponent *StrucComponent) *string { return nil },
		Scope:       Prototype,
		Id:          "IdComponent",
	}
	bike := NewBike()
	bike.Registry(interfaceComponent)
	container, _ := bike.Start()
	_, err := container.InstanceByType((*string)(nil))
	// Then
	if err == nil {
		t.Errorf("InstanceByType must return an error")
	}
}

func TestInstanceByType_GivenComponentWhenInstanceByTypeNoImplementedInterfaceThenReturError(t *testing.T) {
	// Given
	interfaceComponent := Component{
		Type:  (*StrucComponent)(nil),
		Scope: Prototype,
		Id:    "IdComponent",
	}
	bike := NewBike()
	bike.Registry(interfaceComponent)
	container, _ := bike.Start()
	_, err := container.InstanceByType((*NoImplemented)(nil))
	// Then
	if err == nil {
		t.Errorf("InstanceByType must return an error")
	}
}

func TestInstanceByType_GivenComponentConstructorNoReturnPointerWhenRegistryThenReturError(t *testing.T) {
	// Given
	constructorComponent := Component{
		Constructor: NewValueComponent,
		Scope:       Singleton,
	}
	bike := NewBike()
	err := bike.Registry(constructorComponent)
	// Then
	if err == nil {
		t.Errorf("Registry must return an error")
	}
}

func TestRegistry_GivenComponentWithInvalidScopeWhenRegistryThenReturnError(t *testing.T) {
	// Given
	interfaceComponent := Component{
		Constructor: NewInterfaceComponent,
		Scope:       99,
		Id:          "IdComponent",
	}
	bike := NewBike()
	err := bike.Registry(interfaceComponent)
	// Then
	if err == nil {
		t.Errorf("Registry must no return an error")
	}
}
