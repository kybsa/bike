package bike

import "testing"

func Test_GivenCustomScope_WhenGetInstanceByType_ThenReturnNotNil(t *testing.T) {
	// Given
	bike := NewBike()
	err := bike.AddCustomScope(CustomScope, "name")
	if err != nil {
		t.Errorf("AddCustomScope must return nil")
	}
	structComponent := Component{
		Constructor: NewComponent,
		ID:          "IdStructComponent",
		Scope:       CustomScope,
	}
	bike.Add(structComponent)
	container, _ := bike.Start()
	// When
	idContext := "id"
	instance, errInstance := container.InstanceByTypeAndIDContext((*StructComponent)(nil), CustomScope, idContext)
	// Then
	if errInstance != nil {
		t.Errorf("InstanceByTypeAndIDContext must return nil error")

	}
	if instance == nil {
		t.Errorf("InstanceByTypeAndIDContext must return not nil value")
	}
}

func Test_GivenCustomScope_WhenCallGetInstanceByTypeTwoTimes_ThenReturnSameInstances(t *testing.T) {
	// Given
	bike := NewBike()
	err := bike.AddCustomScope(CustomScope, "name")
	if err != nil {
		t.Errorf("AddCustomScope must return nil")
	}
	structComponent := Component{
		Constructor: NewComponent,
		ID:          "IdStructComponent",
		Scope:       CustomScope,
	}
	bike.Add(structComponent)
	container, _ := bike.Start()
	idContext := "id"
	instance1, errInstance1 := container.InstanceByTypeAndIDContext((*StructComponent)(nil), CustomScope, idContext)
	// When
	instance2, errInstance2 := container.InstanceByTypeAndIDContext((*StructComponent)(nil), CustomScope, idContext)

	// Then
	if errInstance1 != nil {
		t.Errorf("InstanceByTypeAndIDContext must return nil error")

	}
	if errInstance2 != nil {
		t.Errorf("InstanceByTypeAndIDContext must return nil error")

	}
	if instance1 == nil {
		t.Errorf("InstanceByTypeAndIDContext must return not nil value")
	}
	if instance1 == instance2 {
		t.Errorf("InstanceByTypeAndIDContext must return same instance")
	}
}

func Test_GivenCustomScope_WhenGetInstanceById_ThenReturnNotNil(t *testing.T) {
	// Given
	bike := NewBike()
	err := bike.AddCustomScope(CustomScope, "name")
	if err != nil {
		t.Errorf("AddCustomScope must return nil")
	}
	structComponent := Component{
		Constructor: NewComponent,
		ID:          "IdStructComponent",
		Scope:       CustomScope,
	}
	bike.Add(structComponent)
	container, _ := bike.Start()
	// When
	idContext := "id"
	instance, errInstance := container.InstanceByIDAndIDContext(structComponent.ID, CustomScope, idContext)
	// Then
	if errInstance != nil {
		t.Errorf("InstanceByIDAndIDContext must return nil error")
	}
	if instance == nil {
		t.Errorf("InstanceByIDAndIDContext must return not nil value")
	}
}

func Test_GivenCustomScope_WhenCallGetInstanceByIdTwoTimes_ThenReturnSameInstances(t *testing.T) {
	// Given
	bike := NewBike()
	err := bike.AddCustomScope(CustomScope, "name")
	if err != nil {
		t.Errorf("AddCustomScope must return nil")
	}
	structComponent := Component{
		Constructor: NewComponent,
		ID:          "IdStructComponent",
		Scope:       CustomScope,
	}
	bike.Add(structComponent)
	container, _ := bike.Start()
	idContext := "id"
	instance1, errInstance1 := container.InstanceByIDAndIDContext(structComponent.ID, CustomScope, idContext)
	// When
	instance2, errInstance2 := container.InstanceByIDAndIDContext(structComponent.ID, CustomScope, idContext)

	// Then
	if errInstance1 != nil {
		t.Errorf("InstanceByIDAndIDContext must return nil error")

	}
	if errInstance2 != nil {
		t.Errorf("InstanceByIDAndIDContext must return nil error")

	}
	if instance1 == nil {
		t.Errorf("InstanceByIDAndIDContext must return not nil value")
	}
	if instance1 == instance2 {
		t.Errorf("InstanceByIDAndIDContext must return same instance")
	}
}

func Test_GivenCustomScope_WhenRemoveContext_ThenReturnNilError(t *testing.T) {
	// Given
	bike := NewBike()
	err := bike.AddCustomScope(CustomScope, "name")
	if err != nil {
		t.Errorf("AddCustomScope must return nil")
	}
	structComponent := Component{
		Constructor: NewComponent,
		ID:          "IdStructComponent",
		Scope:       CustomScope,
	}
	bike.Add(structComponent)
	container, _ := bike.Start()
	idContext := "id"
	_, _ = container.InstanceByIDAndIDContext(structComponent.ID, CustomScope, idContext)
	// When
	errInstance2 := container.RemoveContext(CustomScope, idContext)

	// Then
	if errInstance2 != nil {
		t.Errorf("RemoveContext must return nil error")
	}
}

func Test_GivenInvalidCustomScope_WhenRemoveContext_ThenReturnError(t *testing.T) {
	// Given
	bike := NewBike()
	err := bike.AddCustomScope(CustomScope, "name")
	if err != nil {
		t.Errorf("AddCustomScope must return nil")
	}
	container, _ := bike.Start()
	idContext := "id"
	// When
	errInstance2 := container.RemoveContext(99, idContext)

	// Then
	if errInstance2 == nil {
		t.Errorf("RemoveContext must return an error")
	}
}

func Test_GivenInvalidIdContext_WhenRemoveContext_ThenReturnError(t *testing.T) {
	// Given
	bike := NewBike()
	err := bike.AddCustomScope(CustomScope, "name")
	if err != nil {
		t.Errorf("AddCustomScope must return nil")
	}
	structComponent := Component{
		Constructor: NewComponent,
		ID:          "IdStructComponent",
		Scope:       CustomScope,
	}
	bike.Add(structComponent)
	container, _ := bike.Start()
	idContext := "id"
	container.InstanceByIDAndIDContext(structComponent.ID, CustomScope, idContext)
	// When
	errorRemoveContext := container.RemoveContext(CustomScope, "idContextInvalid")

	// Then
	if errorRemoveContext == nil {
		t.Errorf("RemoveContext must return an error")
	}
}
