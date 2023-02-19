package webtransaction

import (
	"bufio"
	"errors"
	"net"
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

type MockEngine struct {
	CallHandle bool
}

func (mockEngine *MockEngine) Handle(httpMethod, relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	mockEngine.CallHandle = true
	internalResponseWriter := &InternalResponseWriter{}
	context := &gin.Context{
		Writer: internalResponseWriter,
	}
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

type InternalResponseWriter struct {
	CallWrite   bool
	ValueStatus int
}

// Status returns the HTTP response status code of the current request.
func (responseWriter *InternalResponseWriter) Status() int {
	return 1
}

// Size returns the number of bytes already written into the response http body.
// See Written()
func (responseWriter *InternalResponseWriter) Size() int {
	return 1
}

// WriteString writes the string into the response body.
func (responseWriter *InternalResponseWriter) WriteString(string) (int, error) {
	return 1, nil
}

// Written returns true if the response body was already written.
func (responseWriter *InternalResponseWriter) Written() bool {
	return true
}

// WriteHeaderNow forces to write the http header (status code + headers).
func (responseWriter *InternalResponseWriter) WriteHeaderNow() {

}

// Pusher get the http.Pusher for server push
func (responseWriter *InternalResponseWriter) Pusher() http.Pusher {
	return nil
}

func (responseWriter *InternalResponseWriter) CloseNotify() <-chan bool {
	return nil
}

func (responseWriter *InternalResponseWriter) Flush() {

}

func (responseWriter *InternalResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, nil
}

func (responseWriter *InternalResponseWriter) Header() http.Header {
	return http.Header{}
}

func (responseWriter *InternalResponseWriter) Write([]byte) (int, error) {
	responseWriter.CallWrite = true
	return 1, nil
}

func (responseWriter *InternalResponseWriter) WriteHeader(statusCode int) {
	responseWriter.ValueStatus = statusCode
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
	internalResponseWriter := &InternalResponseWriter{}
	context := &gin.Context{
		Writer: internalResponseWriter,
	}
	registryControllerItem := RegistryControllerItem{
		Type: (*RegistryControllerItem)(nil),
	}

	// When
	handRequest(context, registryControllerItem, container)

	// Then
	if !internalResponseWriter.CallWrite {
		t.Errorf("handRequest must call JSON")
	}

	if internalResponseWriter.ValueStatus != http.StatusInternalServerError {
		t.Errorf("handRequest must call JSON with StatusInternalServerError")
	}
}

type Controller struct {
}

func (controller *Controller) Ok(context *gin.Context) (int, interface{}) {
	return http.StatusOK, "val"
}

func (controller *Controller) Error(context *gin.Context) (int, interface{}) {
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
	internalResponseWriter := &InternalResponseWriter{}
	context := &gin.Context{
		Writer: internalResponseWriter,
	}
	registryControllerItem := RegistryControllerItem{
		Type: (*Controller)(nil),
		CallMethod: func(context *gin.Context, inputController interface{}) (int, interface{}) {
			controller := (inputController).(*Controller)
			return controller.Ok(context)
		},
	}

	// When
	handRequest(context, registryControllerItem, container)

	// Then
	if internalResponseWriter.ValueStatus != http.StatusOK {
		t.Errorf("handRequest must call JSON with StatusOK")
	}
	if !internalResponseWriter.CallWrite {
		t.Errorf("handRequest must call JSON")
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
	internalResponseWriter := &InternalResponseWriter{}
	context := &gin.Context{
		Writer: internalResponseWriter,
	}
	registryControllerItem := RegistryControllerItem{
		Type: (*Controller)(nil),
		CallMethod: func(context *gin.Context, inputController interface{}) (int, interface{}) {
			controller := (inputController).(*Controller)
			return controller.Error(context)
		},
	}

	// When
	handRequest(context, registryControllerItem, container)

	// Then
	if !internalResponseWriter.CallWrite {
		t.Errorf("handRequest must call JSON")
	}
	if internalResponseWriter.ValueStatus != http.StatusInternalServerError {
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
	internalResponseWriter := &InternalResponseWriter{}
	context := &gin.Context{
		Writer: internalResponseWriter,
	}
	registryControllerItem := RegistryControllerItem{
		Type: (*Controller)(nil),
		CallMethod: func(context *gin.Context, inputController interface{}) (int, interface{}) {
			controller := (inputController).(*Controller)
			return controller.Ok(context)
		},
	}

	// When
	handRequest(context, registryControllerItem, container)

	// Then
	if !internalResponseWriter.CallWrite {
		t.Errorf("handRequest must call JSON")
	}
	if internalResponseWriter.ValueStatus != http.StatusInternalServerError {
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
	internalResponseWriter := &InternalResponseWriter{}
	context := &gin.Context{
		Writer: internalResponseWriter,
	}
	registryControllerItem := RegistryControllerItem{
		Type: (*Controller)(nil),
		CallMethod: func(context *gin.Context, inputController interface{}) (int, interface{}) {
			controller := (inputController).(*Controller)
			return controller.Error(context)
		},
	}

	// When
	handRequest(context, registryControllerItem, container)

	// Then
	if !internalResponseWriter.CallWrite {
		t.Errorf("handRequest must call JSON")
	}
	if internalResponseWriter.ValueStatus != http.StatusInternalServerError {
		t.Errorf("handRequest must call JSON with StatusInternalServerError")
	}
}
