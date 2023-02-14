package webtransaction

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kybsa/bike"
)

func NewTransactionComponent(gormComponent GormComponent) *TransactionComponent {
	transaction := gormComponent.DB().Begin()
	return &TransactionComponent{
		DB: transaction,
	}
}

func (transactionRequestController *TransactionRequestController) Start(container *bike.Container) {
	for _, item := range transactionRequestController.Items {
		valItem := item
		transactionRequestController.Engine.Handle(valItem.HttMethod, valItem.RelativePath, func(context Context) {
			handRequest(context, valItem, container)
		})
	}
}

func handRequest(context Context, registryControllerItem RegistryControllerItem, container *bike.Container) {
	idRequest := uuid.NewString()
	defer container.RemoveContext(Request, idRequest)
	gormComponentInterface, errTr := container.InstanceByTypeAndIDContext((*GormComponent)(nil), Request, idRequest)
	if errTr != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": errTr.Error()})
		return
	}
	gormComponent := gormComponentInterface.(GormComponent)
	transaction := gormComponent.DB()

	controller, err := container.InstanceByTypeAndIDContext(registryControllerItem.Type, Request, idRequest)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	httpStatus, body := registryControllerItem.CallMethod(context, controller)
	if httpStatus > 199 && httpStatus < 300 {
		transaction.Commit()
		if transaction.Error != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": transaction.Error.Error()})
		} else {
			context.JSON(httpStatus, body)
		}
	} else {
		transaction.Rollback()
		if transaction.Error != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": transaction.Error.Error()})
		} else {
			context.JSON(httpStatus, body)
		}
	}
}

func NewTransactionRequestController(registryController *RegistryController, engine Engine) *TransactionRequestController {
	return &TransactionRequestController{
		Items:  registryController.Items,
		Engine: engine,
	}
}
