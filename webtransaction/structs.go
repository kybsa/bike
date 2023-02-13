package webtransaction

import (
	"github.com/gin-gonic/gin"
	"github.com/kybsa/bike"
	"gorm.io/gorm"
)

const (
	Request bike.Scope = 3
)

type TransactionComponent struct {
	DB *gorm.DB
}

type HandlerFuncController func(context Context, controller interface{}) (int, interface{})

type RegistryControllerItem struct {
	Type         any
	HttMethod    string
	RelativePath string
	CallMethod   HandlerFuncController
}

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
