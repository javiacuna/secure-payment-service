package controller

import (
	"net/http"

	"secure-payment-service/internal/service"
	"secure-payment-service/internal/transfers"

	"github.com/gin-gonic/gin"
)

type TransferController struct {
	transferService service.TransferService
}

func NewTransferController(svc service.TransferService) *TransferController {
	return &TransferController{transferService: svc}
}

func (ctrl *TransferController) CreateTransfer(c *gin.Context) {
	var transfer transfers.TransferRequest
	if err := c.ShouldBindJSON(&transfer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	transferID, err := ctrl.transferService.CreateTransfer(transfer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"transfer_id": transferID,
		"status":      "PENDING",
	})
}

func (ctrl *TransferController) GetTransfer(c *gin.Context) {
	id := c.Param("id")

	transfer, err := ctrl.transferService.GetTransfer(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, transfer)
}

func (ctrl *TransferController) GetAccountBalance(c *gin.Context) {
	id := c.Param("id")

	balance, err := ctrl.transferService.GetAccountBalance(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"account_id": id,
		"balance":    balance,
	})
}

func (ctrl *TransferController) UpdateTransfer(c *gin.Context) {
	var webhook transfers.WebhookEvent
	if err := c.ShouldBindJSON(&webhook); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := ctrl.transferService.UpdateTransfer(webhook.ID, webhook.Status); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "Transfer updated"})
}
