package controller_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"secure-payment-service/internal/controller"
	"secure-payment-service/internal/enums"
	"secure-payment-service/internal/models"
	"secure-payment-service/internal/transfers"
)

const (
	fromAccount     = "acc-001"
	toAccount       = "acc-002"
	amount          = 100.00
	currency        = "USD"
	expTransferID   = "test-transfer-id-123"
	expectedBalance = 750.50
)

func setupRouter(svc *controller.MockTransferService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	ctrl := controller.NewTransferController(svc)

	r.POST("/transfers", ctrl.CreateTransfer)
	r.GET("/transfers/:id", ctrl.GetTransfer)
	r.GET("/accounts/:id/balance", ctrl.GetAccountBalance)
	r.POST("/webhooks/transfer", ctrl.UpdateTransfer)

	return r
}

func givenATransferRequest() transfers.TransferRequest {
	return transfers.TransferRequest{
		FromAccount: fromAccount,
		ToAccount:   toAccount,
		Amount:      amount,
	}
}

func TestCreateTransfer_Success(t *testing.T) {
	svc := controller.NewMockTransferService(t)
	router := setupRouter(svc)

	reqBody := givenATransferRequest()

	svc.EXPECT().CreateTransfer(reqBody).Return(expTransferID, nil).Once()

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/transfers", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusCreated, resp.Code)

	var responseBody map[string]interface{}
	json.Unmarshal(resp.Body.Bytes(), &responseBody)

	assert.Equal(t, expTransferID, responseBody["transfer_id"])
	assert.Equal(t, enums.PENDING.String(), responseBody["status"])
}

func TestCreateTransfer_InvalidJSON(t *testing.T) {
	svc := controller.NewMockTransferService(t)
	router := setupRouter(svc)

	req := httptest.NewRequest(http.MethodPost, "/transfers", bytes.NewBufferString("{invalid json"))
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)

	var responseBody map[string]interface{}
	json.Unmarshal(resp.Body.Bytes(), &responseBody)
	assert.Contains(t, responseBody["error"].(string), "invalid character")
	svc.AssertNotCalled(t, "CreateTransfer", mock.Anything)
}

func TestCreateTransfer_ServiceError(t *testing.T) {
	svc := controller.NewMockTransferService(t)
	router := setupRouter(svc)

	reqBody := givenATransferRequest()
	serviceError := errors.New("error simulado del servicio de transferencia")

	svc.EXPECT().CreateTransfer(reqBody).Return("", serviceError).Once()

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/transfers", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)

	var responseBody map[string]interface{}
	json.Unmarshal(resp.Body.Bytes(), &responseBody)
	assert.Equal(t, serviceError.Error(), responseBody["error"])
}

func TestGetTransfer_Success(t *testing.T) {
	svc := controller.NewMockTransferService(t)
	router := setupRouter(svc)

	expectedTransfer := models.Transfer{
		TransferID:  expTransferID,
		FromAccount: fromAccount,
		ToAccount:   toAccount,
		Amount:      amount,
		Status:      enums.COMPLETED.String(),
	}

	svc.EXPECT().GetTransfer(expTransferID).Return(expectedTransfer, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/transfers/"+expTransferID, nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var responseTransfer models.Transfer
	json.Unmarshal(resp.Body.Bytes(), &responseTransfer)
	assert.Equal(t, expectedTransfer, responseTransfer)
}

func TestGetTransfer_NotFound(t *testing.T) {
	svc := controller.NewMockTransferService(t)
	router := setupRouter(svc)

	serviceError := errors.New("transfer not found")

	svc.EXPECT().GetTransfer(expTransferID).Return(models.Transfer{}, serviceError).Once()

	req := httptest.NewRequest(http.MethodGet, "/transfers/"+expTransferID, nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusNotFound, resp.Code)

	var responseBody map[string]interface{}
	json.Unmarshal(resp.Body.Bytes(), &responseBody)
	assert.Equal(t, serviceError.Error(), responseBody["error"])
}

func TestGetAccountBalance_Success(t *testing.T) {
	svc := controller.NewMockTransferService(t)
	router := setupRouter(svc)

	svc.EXPECT().GetAccountBalance(fromAccount).Return(expectedBalance, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/accounts/"+fromAccount+"/balance", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var responseBody map[string]interface{}
	json.Unmarshal(resp.Body.Bytes(), &responseBody)
	assert.Equal(t, fromAccount, responseBody["account_id"])
	assert.Equal(t, expectedBalance, responseBody["balance"])
}

func TestGetAccountBalance_ServiceError(t *testing.T) {
	svc := controller.NewMockTransferService(t)
	router := setupRouter(svc)

	serviceError := errors.New("error obteniendo balance")

	svc.EXPECT().GetAccountBalance(fromAccount).Return(0.0, serviceError).Once()

	req := httptest.NewRequest(http.MethodGet, "/accounts/"+fromAccount+"/balance", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusNotFound, resp.Code)
	var responseBody map[string]interface{}
	json.Unmarshal(resp.Body.Bytes(), &responseBody)
	assert.Equal(t, serviceError.Error(), responseBody["error"])
}

func givenAWebhookEvent() transfers.WebhookEvent {
	return transfers.WebhookEvent{
		ID:     expTransferID,
		Status: enums.COMPLETED.String(),
	}
}

func TestUpdateTransfer_Success(t *testing.T) {
	svc := controller.NewMockTransferService(t)
	router := setupRouter(svc)

	webhookBody := givenAWebhookEvent()

	svc.EXPECT().UpdateTransfer(webhookBody.ID, webhookBody.Status).Return(nil).Once()

	jsonBody, _ := json.Marshal(webhookBody)
	req := httptest.NewRequest(http.MethodPost, "/webhooks/transfer", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var responseBody map[string]interface{}
	json.Unmarshal(resp.Body.Bytes(), &responseBody)
	assert.Equal(t, "Transfer updated", responseBody["status"])
}

func TestUpdateTransfer_InvalidJSON(t *testing.T) {
	svc := controller.NewMockTransferService(t)
	router := setupRouter(svc)

	req := httptest.NewRequest(http.MethodPost, "/webhooks/transfer", bytes.NewBufferString("{invalid json"))
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
	var responseBody map[string]interface{}
	json.Unmarshal(resp.Body.Bytes(), &responseBody)
	assert.Contains(t, responseBody["error"].(string), "invalid character")
	svc.AssertNotCalled(t, "UpdateTransfer", mock.Anything, mock.Anything)
}

func TestUpdateTransfer_ServiceError(t *testing.T) {
	svc := controller.NewMockTransferService(t)
	router := setupRouter(svc)

	webhookBody := givenAWebhookEvent()
	serviceError := errors.New("transfer not found")

	svc.EXPECT().UpdateTransfer(webhookBody.ID, webhookBody.Status).Return(serviceError).Once()

	jsonBody, _ := json.Marshal(webhookBody)
	req := httptest.NewRequest(http.MethodPost, "/webhooks/transfer", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusNotFound, resp.Code)
	var responseBody map[string]interface{}
	json.Unmarshal(resp.Body.Bytes(), &responseBody)
	assert.Equal(t, serviceError.Error(), responseBody["error"])
}
