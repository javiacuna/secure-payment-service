package routes

import (
	"secure-payment-service/internal/controller"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine, authMiddleware gin.HandlerFunc, transferCtrl *controller.TransferController) {
	v1 := router.Group("/api/v1")
	v1.Use(authMiddleware)
	v1.POST("/transfer", transferCtrl.CreateTransfer)
	v1.GET("/transfer/:id", transferCtrl.GetTransfer)
	v1.GET("/account/:id/balance", transferCtrl.GetAccountBalance)

	router.POST("/api/v1/webhook", transferCtrl.UpdateTransfer)
	router.GET("/metrics", controller.PrometheusHandler())
}
