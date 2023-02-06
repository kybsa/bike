package bike

import (
	"reflect"
	"testing"
	"time"
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

type StructComponent struct {
	InitStatus   bool
	PostStart    bool
	StopStatus   bool
	SetContainer bool
}

func (_self *StructComponent) DoAnything() {

}

func (_self *StructComponent) Init() {
	_self.InitStatus = true
}

func (_self *StructComponent) InitReturnError() error {
	return &Error{}
}

func (_self *StructComponent) PostInit() {
	_self.PostStart = true
}

func (_self *StructComponent) PostInitWithContainer(container *Container) {
	_self.SetContainer = true
}

func (_self *StructComponent) InvalidPostInit(param string) {
}

func (_self *StructComponent) InvalidPostInitTwoArguments(param1 string, param2 int) {
}

func (_self *StructComponent) InvalidInit(param string) {
}

func (_self *StructComponent) Stop() {
	_self.StopStatus = true
}

func (_self *StructComponent) StopError() error {
	return &Error{}
}

func (_self *StructComponent) DoAny() {
}

func NewInterfaceComponent() InterfaceComponent {
	return &StructComponent{}
}

func NewComponent() *StructComponent {
	return &StructComponent{
		InitStatus: false,
		PostStart:  false,
		StopStatus: false,
	}
}

func NewValueComponent() StructComponent {
	return StructComponent{}
}

func InvalidConstructor() {
}

func NewComponentReturnError() (*StructComponent, error) {
	return &StructComponent{}, &Error{messageError: "message"}
}

func NewComponentReturnNilError() (*StructComponent, error) {
	return &StructComponent{}, nil
}

func NewComponentReturnNoError() (*StructComponent, *StructComponent) {
	return &StructComponent{}, nil
}

type A struct {
}

type B struct {
	a *A
}

func NewA() *A {
	return &A{}
}

func NewB(a *A) *B {
	return &B{a: a}
}

func TestInstanceByID_GivenComponentWithId_WhenInstanceByID_ThenReturnNotNill(t *testing.T) {
	// Given
	structComponent := Component{
		Constructor: NewComponent,
		ID:          "IdStructComponent",
	}
	bike := NewBike()
	bike.Add(structComponent)
	container, _ := bike.Start()
	// When
	instance, err := container.InstanceByID(structComponent.ID)
	// Then
	if err != nil {
		t.Errorf("Error to get instance by id:%s.", err.Error())
	}
	if instance == nil {
		t.Errorf("InstanceById return nil.")
	}
}

func TestInstanceByType_GivenComponentWithTypeAndInterfaces_WhenInstanceByType_ThenReturnNotNill(t *testing.T) {
	// Given
	structComponent := Component{
		Constructor: NewComponent,
		Interfaces:  []any{(*InterfaceComponent)(nil)},
	}
	bike := NewBike()
	bike.Add(structComponent)
	container, _ := bike.Start()
	// When
	instance, err := container.InstanceByType((*InterfaceComponent)(nil))
	// Then
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
		Constructor:   NewComponent,
		Scope:         Prototype,
		PostConstruct: "Init",
	}
	bike := NewBike()
	bike.Add(structComponent)
	container, _ := bike.Start()
	instance, _ := container.InstanceByType((*StructComponent)(nil))
	StructComponentInstance := (instance).(*StructComponent)
	// When
	stopErr := container.Stop()
	if stopErr != nil {
		t.Errorf("Error to Stop :%s.", stopErr.Error())
	}

	// Then
	time.Sleep(100 * time.Millisecond)
	if StructComponentInstance.InitStatus == false {
		t.Errorf("Bike doesn't call init method StructComponent")
	}
}

func TestStop_GivenComponentWithScopePrototype_WhenStop_ThenCallDestroyMethod(t *testing.T) {
	// Given
	structComponent := Component{
		Constructor: NewComponent,
		Scope:       Prototype,
		Destroy:     "Stop",
	}
	bike := NewBike()
	bike.Add(structComponent)
	container, _ := bike.Start()
	instance, _ := container.InstanceByType((*StructComponent)(nil))
	StructComponentInstance := (instance).(*StructComponent)
	StructComponentInstance.StopStatus = false
	// When
	stopErr := container.Stop()
	if stopErr != nil {
		t.Errorf("Error to Stop :%s.", stopErr.Error())
	}
	// Then
	time.Sleep(100 * time.Millisecond)
	if StructComponentInstance.StopStatus == false {
		t.Errorf("Bike doesn't call stop method StructComponent")
	}
}

func TestStop_GivenComponentWithScopeSingleton_WhenStop_ThenCallDestroyMethod(t *testing.T) {
	// Given
	structComponent := Component{
		Constructor: NewComponent,
		Scope:       Singleton,
		Destroy:     "Stop",
	}
	bike := NewBike()
	bike.Add(structComponent)
	container, _ := bike.Start()
	instance, _ := container.InstanceByType((*StructComponent)(nil))
	StructComponentInstance := (instance).(*StructComponent)
	StructComponentInstance.StopStatus = false
	// When
	stopErr := container.Stop()
	if stopErr != nil {
		t.Errorf("Error to Stop :%s.", stopErr.Error())
	}
	// Then
	if StructComponentInstance.StopStatus == false {
		t.Errorf("Bike doesn't call stop method StructComponent")
	}
}

func TestInstanceById_GivenComponentWithConstructorAndIdAndScopeSingleton_WhenRegistry_ThenReturnNotNullInstanceById(t *testing.T) {
	// Given
	interfaceComponent := Component{
		Constructor: NewInterfaceComponent,
		Scope:       Singleton,
		ID:          "IdComponent",
	}
	bike := NewBike()
	// When
	bike.Add(interfaceComponent)
	container, _ := bike.Start()
	// Then
	instance, err := container.InstanceByID(interfaceComponent.ID)
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
		ID:          "IdComponent",
	}
	bike := NewBike()
	bike.Add(interfaceComponent)
	container, _ := bike.Start()
	// When
	instance, err := container.InstanceByID(interfaceComponent.ID)
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
		Constructor: NewComponent,
		Scope:       Prototype,
		ID:          "IdComponent",
	}
	bike := NewBike()
	bike.Add(structComponent)
	container, _ := bike.Start()
	// When
	instance, err := container.InstanceByID(structComponent.ID)
	// Then
	if err != nil {
		t.Errorf("Error to get instance by id:%s.", err.Error())
	}
	if instance == nil {
		t.Errorf("InstanceById return nil.")
	}
}

func TestStart_GivenComponentWithInvalidConstructor_WhenStart_ThenReturnError(t *testing.T) {
	// Given
	structComponent := Component{
		Constructor: InvalidConstructor,
		Scope:       Prototype,
		ID:          "IdComponent",
	}
	bike := NewBike()
	bike.Add(structComponent)
	// When
	_, startErr := bike.Start()
	// Then
	if startErr == nil {
		t.Errorf("Start must return an error")
	}
}

func TestInstanceById_GivenComponentWithInvalidScope_WhenRegistry_ThenReturnError(t *testing.T) {
	// Given
	structComponent := Component{
		Constructor: InvalidConstructor,
		Scope:       99,
		ID:          "IdComponent",
	}
	bike := NewBike()
	bike.Add(structComponent)
	// When
	_, startErr := bike.Start()
	// Then
	if startErr == nil {
		t.Errorf("InstanceById must return an error ")
	}
}

func TestInstanceById_GivenComponentWithInvalidId_WhenInstanceById_ThenReturnError(t *testing.T) {
	// Given
	bike := NewBike()
	container, _ := bike.Start()
	// When
	_, err := container.InstanceByID("any-id")
	// Then
	if err == nil {
		t.Errorf("InstanceById must return an error ")
	}
}

func TestInstanceByType_GivenComponentImplementInterface_WhenInstanceByType_ThenReturnInstance(t *testing.T) {
	// Given
	interfaceComponent := Component{
		Constructor: NewInterfaceComponent,
		Scope:       Prototype,
		ID:          "IdComponent",
	}
	bike := NewBike()
	bike.Add(interfaceComponent)

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

func TestInstanceByType_GivenComponentScopePrototypeWithInvalidConstructorWhenInstanceByTypeThenReturnError(t *testing.T) {
	// Given
	interfaceComponent := Component{
		Constructor: func(StructComponent *StructComponent) *string { return nil },
		Scope:       Prototype,
		ID:          "IdComponent",
	}
	bike := NewBike()
	bike.Add(interfaceComponent)
	container, _ := bike.Start()
	_, err := container.InstanceByType((*string)(nil))
	// Then
	if err == nil {
		t.Errorf("InstanceByType must return an error")
	}
}

func TestInstanceByType_GivenComponentWhenInstanceByTypeNoImplementedInterfaceThenReturnError(t *testing.T) {
	// Given
	interfaceComponent := Component{
		Constructor: NewInterfaceComponent,
		Scope:       Prototype,
		ID:          "IdComponent",
	}
	bike := NewBike()
	bike.Add(interfaceComponent)
	container, _ := bike.Start()
	_, err := container.InstanceByType((*NoImplemented)(nil))
	// Then
	if err == nil {
		t.Errorf("InstanceByType must return an error")
	}
}

func TestStart_GivenComponentConstructorNoReturnPointerWhenRegistryThenReturnError(t *testing.T) {
	// Given
	constructorComponent := Component{
		Constructor: NewValueComponent,
		Scope:       Singleton,
	}
	bike := NewBike()
	bike.Add(constructorComponent)
	// When
	_, err := bike.Start()
	// Then
	if err == nil {
		t.Errorf("Start must return an error")
	}
}

func TestStart_GivenComponentNullTypeAndConstructor_WhenStart_ThenReturnError(t *testing.T) {
	// Given
	component := Component{
		Scope: Singleton,
	}
	// When
	bike := NewBike()
	bike.Add(component)
	// When
	_, err := bike.Start()
	// Then
	if err == nil {
		t.Errorf("Start must return an error")
	}
}

func TestInstanceByType_GivenComponentWithADependency_WhenInstanceByType_ThenReturnInstance(t *testing.T) {
	// Given
	bike := NewBike()
	componentA := Component{
		Constructor: NewA,
		Scope:       Singleton,
	}
	bike.Add(componentA)

	componentB := Component{
		Constructor: NewB,
		Scope:       Singleton,
	}
	bike.Add(componentB)

	container, _ := bike.Start()
	// When
	_, err := container.InstanceByType((*B)(nil))
	// Then
	if err != nil {
		t.Errorf("InstanceByType must no return an error")
	}
}

func TestStart_GivenComponentInvalidPostConstructName_WhenStart_ThenReturnError(t *testing.T) {
	// Given
	component := Component{
		Constructor:   NewComponent,
		Scope:         Singleton,
		PostConstruct: "StartInit",
	}
	bike := NewBike()
	bike.Add(component)
	// When
	_, err := bike.Start()
	// Then
	if err == nil {
		t.Errorf("Start must return an error")
	}
}

func TestStart_GivenComponentInvalidPostConstruct_WhenStart_ThenReturnError(t *testing.T) {
	// Given
	component := Component{
		Constructor:   NewComponent,
		Scope:         Singleton,
		PostConstruct: "InvalidInit",
	}
	bike := NewBike()
	bike.Add(component)
	// When
	_, err := bike.Start()
	// Then
	if err == nil {
		t.Errorf("Start must return an error")
	}
}

func TestRegistry_GivenComponentInvalidDestroyName_WhenStart_ThenReturnError(t *testing.T) {
	// Given
	component := Component{
		Constructor: NewComponent,
		Scope:       Singleton,
		Destroy:     "StartInit",
	}
	bike := NewBike()
	bike.Add(component)
	// When
	_, err := bike.Start()
	// Then
	if err == nil {
		t.Errorf("Start must return an error")
	}
}

func TestRegistry_GivenComponentInvalidDestroy_WhenStart_ThenReturnError(t *testing.T) {
	// Given
	component := Component{
		Constructor: NewComponent,
		Scope:       Singleton,
		Destroy:     "InvalidInit",
	}
	bike := NewBike()
	bike.Add(component)
	// When
	_, err := bike.Start()
	// Then
	if err == nil {
		t.Errorf("Start must return an error")
	}
}

func TestInstanceById_GivenComponentInvalidDependenciesWhenInstanceByIdThenReturnError(t *testing.T) {
	// Given
	component := Component{
		Constructor: NewB,
		Scope:       Prototype,
		ID:          "IdComponent",
	}

	bike := NewBike()
	bike.Add(component)
	// When
	container, _ := bike.Start()
	// Then
	_, err := container.InstanceByID(component.ID)
	if err == nil {
		t.Errorf("InstanceById must return an error")
	}
}

func TestInstanceById_GivenComponentInvalidDependenciesWhenStartIdThenReturnError(t *testing.T) {
	// Given
	component := Component{
		Constructor: NewB,
		Scope:       Singleton,
		ID:          "IdComponent",
	}
	// When
	bike := NewBike()
	bike.Add(component)

	_, err := bike.Start()
	if err == nil {
		t.Errorf("InstanceById must return an error")
	}
}

func TestStart_GivenComponentConstructorReturnError_WhenStart_ThenReturnError(t *testing.T) {
	// Given
	component := Component{
		Constructor: NewComponentReturnError,
		Scope:       Singleton,
	}
	bike := NewBike()
	bike.Add(component)
	// When
	_, err := bike.Start()
	// Then
	if err == nil {
		t.Errorf("Start must return an error")
	}
}

func TestStart_GivenComponentConstructorNoReturn_WhenStart_ThenReturnNilError(t *testing.T) {
	// Given
	component := Component{
		Constructor: NewComponentReturnNilError,
		Scope:       Singleton,
	}
	bike := NewBike()
	bike.Add(component)
	// When
	_, err := bike.Start()
	// Then
	if err != nil {
		t.Errorf("Start must return nil error")
	}

}

func TestStart_GivenComponentConstructorReturnNoError_WhenStart_ThenReturnError(t *testing.T) {
	// Given
	component := Component{
		Constructor: NewComponentReturnNoError,
		Scope:       Singleton,
	}
	bike := NewBike()
	bike.Add(component)
	// When
	_, err := bike.Start()
	// Then
	if err == nil {
		t.Errorf("Start must return an error")
	}
}

func Test_GivenInterfaceType_WhenGetTypeName_ThenReturnDo(t *testing.T) {
	// Given
	type Do interface {
	}
	anyDo := (*Do)(nil)
	interfaceType := reflect.TypeOf(anyDo).Elem()

	// When
	actual := getTypeName(interfaceType)

	// Then
	if actual != "Do" {
		t.Errorf("getTypeName must Do")
	}
}

func Test_GivenPointerPointerType_WhenGetTypeName_ThenReturnPointerPointerDo(t *testing.T) {
	// Given
	type Do interface {
	}
	anyDo := (**Do)(nil)
	interfaceType := reflect.TypeOf(anyDo)

	// When
	actual := getTypeName(interfaceType)

	// Then
	if actual != "**Do" {
		t.Errorf("getTypeName *DoDo, current value:%s", actual)
	}
}

func Test_GivenStructPointerType_WhenGetTypeName_ThenReturnDo(t *testing.T) {
	// Given
	type Str struct {
	}
	anyDo := (*Str)(nil)
	interfaceType := reflect.TypeOf(anyDo)

	// When
	actual := getTypeName(interfaceType)

	// Then
	if actual != "*Str" {
		t.Errorf("getTypeName must return an *Str, current value:%s", actual)
	}
}

func TestStart_GivenComponentWithPostStart_WhenStart_ThenCallPostInitMethod(t *testing.T) {
	// Given
	structComponent := Component{
		Constructor: NewComponent,
		Scope:       Singleton,
		PostStart:   "PostInit",
	}
	bike := NewBike()
	bike.Add(structComponent)
	// When
	container, _ := bike.Start()
	instance, _ := container.InstanceByType((*StructComponent)(nil))
	StructComponentInstance := (instance).(*StructComponent)
	// Then
	time.Sleep(200 * time.Millisecond)
	if StructComponentInstance.PostStart == false {
		t.Errorf("Bike doesn't call init method StructComponent")
	}
}

func TestStart_GivenComponentWithInvalidNamePostStart_WhenStart_ThenReturnError(t *testing.T) {
	// Given
	structComponent := Component{
		Constructor: NewComponent,
		Scope:       Singleton,
		PostStart:   "InvalidNameMethod",
	}
	bike := NewBike()
	bike.Add(structComponent)
	_, startError := bike.Start()
	if startError == nil {
		t.Errorf("Start must return an error")
	}
}

func TestStart_GivenComponentWithInvalidPostStart_WhenStart_ThenReturnError(t *testing.T) {
	// Given
	structComponent := Component{
		Constructor: NewComponent,
		Scope:       Singleton,
		PostStart:   "InvalidPostInit",
	}
	bike := NewBike()
	bike.Add(structComponent)
	_, startError := bike.Start()
	if startError == nil {
		t.Errorf("Start must return an error")
	}
}

func TestStart_GivenComponentWithInvalidPostStartTwoArguments_WhenStart_ThenReturnError(t *testing.T) {
	// Given
	structComponent := Component{
		Constructor:   NewComponent,
		Scope:         Singleton,
		PostConstruct: "InvalidPostInitTwoArguments",
	}
	bike := NewBike()
	bike.Add(structComponent)
	_, startError := bike.Start()
	if startError == nil {
		t.Errorf("Start must return an error")
	}
}

func TestStart_GivenComponentWithPostStartAndScopePrototype_WhenStart_ThenReturnError(t *testing.T) {
	// Given
	structComponent := Component{
		Constructor: NewComponent,
		Scope:       Prototype,
		PostStart:   "PostInit",
	}
	bike := NewBike()
	bike.Add(structComponent)
	// When
	_, startError := bike.Start()
	if startError == nil {
		t.Errorf("Start must return an error")
	}
}

func TestStart_GivenComponentWithInitReturnError_WhenStart_ThenReturnError(t *testing.T) {
	// Given
	structComponent := Component{
		Constructor:   NewComponent,
		Scope:         Singleton,
		PostConstruct: "InitReturnError",
	}
	bike := NewBike()
	bike.Add(structComponent)
	// When
	_, startErr := bike.Start()
	// Then
	if startErr == nil {
		t.Errorf("Start must return an error")
	}
}

func TestStart_GivenComponentSingletonStopReturnError_WhenStop_ThenReturnError(t *testing.T) {
	// Given
	structComponent := Component{
		Constructor: NewComponent,
		Scope:       Singleton,
		Destroy:     "StopError",
	}
	bike := NewBike()
	bike.Add(structComponent)
	// When
	container, _ := bike.Start()
	stopErr := container.Stop()
	// Then
	if stopErr == nil {
		t.Errorf("Stop must return an error")
	}
}

func TestStart_GivenComponentPrototypeStopReturnError_WhenStop_ThenReturnError(t *testing.T) {
	// Given
	structComponent := Component{
		ID:          "1",
		Constructor: NewComponent,
		Scope:       Prototype,
		Destroy:     "StopError",
	}
	bike := NewBike()
	bike.Add(structComponent)
	// When
	container, _ := bike.Start()
	_, err := container.InstanceByID("1")
	if err != nil {
		t.Errorf("InstanceByID must return no error")
	}
	stopErr := container.Stop()
	// Then
	if stopErr == nil {
		t.Errorf("Stop must return an error")
	}
}

func TestStart_GivenComponentWithPostConstructArgument_WhenStart_ThenPostConstructorReceiveContainerArgument(t *testing.T) {
	// Given
	structComponent := Component{
		ID:            "1",
		Constructor:   NewComponent,
		Scope:         Prototype,
		PostConstruct: "PostInitWithContainer",
	}
	bike := NewBike()
	bike.Add(structComponent)
	// When
	container, startErr := bike.Start()
	if startErr != nil {
		t.Errorf("Start must return no error, Current error:%s", startErr.Error())
		return
	}

	interfaceComponent, err := container.InstanceByID("1")
	if err != nil {
		t.Errorf("InstanceByID must return no error")
	}
	structValue := interfaceComponent.(*StructComponent)
	// Then
	if structValue.SetContainer == false {
		t.Errorf("Start must call PostConstruct setting the container on argument")
	}
}

const (
	CustomScope Scope = 2
)

func Test_GivenValidCustomScopeWhenAddCustomScopeThenReturnNill(t *testing.T) {
	// Given
	bike := NewBike()
	// When
	err := bike.AddCustomScope(CustomScope, "name")
	// Then
	if err != nil {
		t.Errorf("AddCustomScope must return no error")
	}
}

func Test_GivenSingletonScopeWhenAddCustomScopeThenReturnError(t *testing.T) {
	// Given
	bike := NewBike()
	// When
	err := bike.AddCustomScope(Singleton, "name")
	// Then
	if err == nil {
		t.Errorf("AddCustomScope must return an error")
	}
}

func Test_GivenCustomScopeExistWhenAddCustomScopeThenReturnNill(t *testing.T) {
	// Given
	bike := NewBike()
	errCustomScope := bike.AddCustomScope(CustomScope, "name")
	if errCustomScope != nil {
		t.Errorf("AddCustomScope must return nil error")
	}
	// When
	err := bike.AddCustomScope(CustomScope, "name")
	// Then
	if err == nil {
		t.Errorf("AddCustomScope must return an error")
	}
}
