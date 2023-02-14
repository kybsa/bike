package webtransaction

import (
	"errors"
	"net/http"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
	"github.com/kybsa/bike"
)

type MockDBComponent struct {
	db *gorm.DB
}

func (mockDbComponent *MockDBComponent) DB() *gorm.DB {
	return mockDbComponent.db
}

func NewMockDBComponent() *MockDBComponent {
	gdb, _ := gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{})
	return &MockDBComponent{
		db: gdb.Begin(),
	}
}

func TestNewTransactionPostgresComponent_GivenPostgresComponent_WhenNewTransactionPostgresComponentThenReturnNotNull(t *testing.T) {
	// Given
	gdb, _ := gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{})
	postgresComponent := &MockDBComponent{
		db: gdb,
	}
	// When
	transactionPostgresComponent := NewTransactionComponent(postgresComponent)
	// Then
	if transactionPostgresComponent == nil {
		t.Errorf("NewTransactionPostgresComponent must return not nil value")
	}
}

type MockContext struct {
	CallJSON bool
	Code     int
	Body     any
}

func (context *MockContext) JSON(code int, obj any) {
	context.CallJSON = true
	context.Code = code
	context.Body = obj
}

type MockEngine struct {
	CallHandle bool
}

func (mockEngine *MockEngine) Handle(httpMethod, relativePath string, handlers ...HandlerFunc) gin.IRoutes {
	mockEngine.CallHandle = true
	context := &MockContext{}
	for _, handler := range handlers {
		handler(context)
	}
	return nil
}

func TestStart_GivenTransactionRequestController_WhenStart_ThenCallHandle(t *testing.T) {
	// Given
	registryController := &RegistryController{
		Items: []RegistryControllerItem{
			{HttMethod: "POST", RelativePath: "/path"},
		},
	}
	engine := &MockEngine{}
	container := bike.Container{}
	transactionRequestController := NewTransactionRequestController(registryController, engine)
	// When
	transactionRequestController.Start(&container)
	// Then
	if !engine.CallHandle {
		t.Errorf("Start must call handle method")
	}
}

func Test_GivenInvalidController_WhenHandRequest_ThenCallJSONWithInternalServerError(t *testing.T) {
	// Given
	bk := bike.NewBike()
	errCustomScope := bk.AddCustomScope(Request, "Request")
	if errCustomScope != nil {
		t.Errorf("AddCustomScope mus return nil error")
	}
	// OK
	bk.Add(bike.Component{
		Constructor: NewMockDBComponent,
		Scope:       bike.Singleton,
		Interfaces:  []any{(*GormComponent)(nil)},
	})
	bk.Add(bike.Component{
		ID:          "NewTransactionPostgresComponent",
		Constructor: NewTransactionComponent,
		Scope:       Request,
	})
	container, errStartBike := bk.Start()
	if errStartBike != nil {
		t.Errorf("Start must no return nil. Error:[%s]", errStartBike.Error())
	}
	context := &MockContext{}
	registryControllerItem := RegistryControllerItem{
		Type: (*RegistryControllerItem)(nil),
	}

	// When
	handRequest(context, registryControllerItem, container)

	// Then
	if !context.CallJSON {
		t.Errorf("handRequest must call JSON method")
	}

	if context.Code != http.StatusInternalServerError {
		t.Errorf("handRequest must call JSON with StatusInternalServerError")
	}
}

type Controller struct {
}

func (controller *Controller) Ok(context Context) (int, interface{}) {
	return http.StatusOK, "val"
}

func (controller *Controller) Error(context Context) (int, interface{}) {
	return http.StatusInternalServerError, "val"
}

func NewController() *Controller {
	return &Controller{}
}

func Test_GivenControllerReturnOk_WhenHandRequest_ThenCallJSONWithOkStatus(t *testing.T) {
	// Given
	bk := bike.NewBike()
	errCustomScope := bk.AddCustomScope(Request, "Request")
	if errCustomScope != nil {
		t.Errorf("AddCustomScope mus return nil error")
	}
	bk.Add(bike.Component{
		Constructor: NewMockDBComponent,
		Scope:       bike.Singleton,
		Interfaces:  []any{(*GormComponent)(nil)},
	})
	bk.Add(bike.Component{
		ID:          "NewTransactionComponent",
		Constructor: NewTransactionComponent,
		Scope:       Request,
	})
	bk.Add(bike.Component{
		ID:          "NewController",
		Constructor: NewController,
		Scope:       Request,
	})
	container, errStartBike := bk.Start()
	if errStartBike != nil {
		t.Errorf("Start must no return nil. Error:[%s]", errStartBike.Error())
	}
	context := &MockContext{}
	registryControllerItem := RegistryControllerItem{
		Type: (*Controller)(nil),
		CallMethod: func(context Context, inputController interface{}) (int, interface{}) {
			controller := (inputController).(*Controller)
			return controller.Ok(context)
		},
	}

	// When
	handRequest(context, registryControllerItem, container)

	// Then
	if !context.CallJSON {
		t.Errorf("handRequest must call JSON method")
	}

	if context.Code != http.StatusOK {
		t.Errorf("handRequest must call JSON with StatusOK")
	}
}

func Test_GivenControllerReturnError_WhenHandRequest_ThenCallJSONWithInternalServerError(t *testing.T) {
	// Given
	bk := bike.NewBike()
	errCustomScope := bk.AddCustomScope(Request, "Request")
	if errCustomScope != nil {
		t.Errorf("AddCustomScope mus return nil error")
	}
	bk.Add(bike.Component{
		Constructor: NewMockDBComponent,
		Scope:       bike.Singleton,
		Interfaces:  []any{(*GormComponent)(nil)},
	})
	bk.Add(bike.Component{
		ID:          "NewTransactionComponent",
		Constructor: NewTransactionComponent,
		Scope:       Request,
	})
	bk.Add(bike.Component{
		ID:          "NewController",
		Constructor: NewController,
		Scope:       Request,
	})
	container, errStartBike := bk.Start()
	if errStartBike != nil {
		t.Errorf("Start must no return nil. Error:[%s]", errStartBike.Error())
	}
	context := &MockContext{}
	registryControllerItem := RegistryControllerItem{
		Type: (*Controller)(nil),
		CallMethod: func(context Context, inputController interface{}) (int, interface{}) {
			controller := (inputController).(*Controller)
			return controller.Error(context)
		},
	}

	// When
	handRequest(context, registryControllerItem, container)

	// Then
	if !context.CallJSON {
		t.Errorf("handRequest must call JSON method")
	}

	if context.Code != http.StatusInternalServerError {
		t.Errorf("handRequest must call JSON with StatusInternalServerError")
	}
}

func NewMockDBComponentTransactionFail() *MockDBComponent {
	db, _ := gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{})
	tx := db.Begin()
	tx.Error = errors.New("error")
	return &MockDBComponent{
		db: tx,
	}
}

func Test_GivenControllerReturnOkAndTransactionFail_WhenHandRequest_ThenCallJSONWithInternalServerError(t *testing.T) {
	// Given
	bk := bike.NewBike()
	errCustomScope := bk.AddCustomScope(Request, "Request")
	if errCustomScope != nil {
		t.Errorf("AddCustomScope mus return nil error")
	}
	bk.Add(bike.Component{
		Constructor: NewMockDBComponentTransactionFail,
		Scope:       bike.Singleton,
		Interfaces:  []any{(*GormComponent)(nil)},
	})
	bk.Add(bike.Component{
		ID:          "NewTransactionComponent",
		Constructor: NewTransactionComponent,
		Scope:       Request,
	})
	bk.Add(bike.Component{
		ID:          "NewController",
		Constructor: NewController,
		Scope:       Request,
	})
	container, errStartBike := bk.Start()
	if errStartBike != nil {
		t.Errorf("Start must no return nil. Error:[%s]", errStartBike.Error())
	}
	context := &MockContext{}
	registryControllerItem := RegistryControllerItem{
		Type: (*Controller)(nil),
		CallMethod: func(context Context, inputController interface{}) (int, interface{}) {
			controller := (inputController).(*Controller)
			return controller.Ok(context)
		},
	}

	// When
	handRequest(context, registryControllerItem, container)

	// Then
	if !context.CallJSON {
		t.Errorf("handRequest must call JSON method")
	}

	if context.Code != http.StatusInternalServerError {
		t.Errorf("handRequest must call JSON with StatusInternalServerError")
	}
}

func Test_GivenControllerReturnErrorAndTransactionFail_WhenHandRequest_ThenCallJSONWithInternalServerError(t *testing.T) {
	// Given
	bk := bike.NewBike()
	errCustomScope := bk.AddCustomScope(Request, "Request")
	if errCustomScope != nil {
		t.Errorf("AddCustomScope mus return nil error")
	}
	bk.Add(bike.Component{
		Constructor: NewMockDBComponentTransactionFail,
		Scope:       bike.Singleton,
		Interfaces:  []any{(*GormComponent)(nil)},
	})
	bk.Add(bike.Component{
		ID:          "NewTransactionComponent",
		Constructor: NewTransactionComponent,
		Scope:       Request,
	})
	bk.Add(bike.Component{
		ID:          "NewController",
		Constructor: NewController,
		Scope:       Request,
	})
	container, errStartBike := bk.Start()
	if errStartBike != nil {
		t.Errorf("Start must no return nil. Error:[%s]", errStartBike.Error())
	}
	context := &MockContext{}
	registryControllerItem := RegistryControllerItem{
		Type: (*Controller)(nil),
		CallMethod: func(context Context, inputController interface{}) (int, interface{}) {
			controller := (inputController).(*Controller)
			return controller.Error(context)
		},
	}

	// When
	handRequest(context, registryControllerItem, container)

	// Then
	if !context.CallJSON {
		t.Errorf("handRequest must call JSON method")
	}

	if context.Code != http.StatusInternalServerError {
		t.Errorf("handRequest must call JSON with StatusInternalServerError")
	}
}
