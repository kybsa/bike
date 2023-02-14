// Package webtransaction to implement transaction on a web applications
package webtransaction

import (
	"github.com/gin-gonic/gin"
	"github.com/kybsa/bike"
	"gorm.io/gorm"
)

const (
	// Request scope to use on web development
	Request bike.Scope = 3
)

// TransactionComponent store a DB with begin transaction
type TransactionComponent struct {
	DB *gorm.DB
}

// HandlerFuncController define interface function to call methods on registry controller config
type HandlerFuncController func(context Context, controller interface{}) (int, interface{})

// RegistryControllerItem store dato to registry a controller
type RegistryControllerItem struct {
	Type         any
	HttMethod    string
	RelativePath string
	CallMethod   HandlerFuncController
}

// RegistryController Store array of RegistryControllerItem
type RegistryController struct {
	Items []RegistryControllerItem
}

type Context interface {
	JSON(code int, obj any)
}

type HandlerFunc func(Context)

type Engine interface {
	Handle(httpMethod, relativePath string, handlers ...HandlerFunc) gin.IRoutes
}

type TransactionRequestController struct {
	Items  []RegistryControllerItem
	Engine Engine
}

type GormComponent interface {
	DB() *gorm.DB
}
