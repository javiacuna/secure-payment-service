package service_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"secure-payment-service/internal/models"
	"secure-payment-service/internal/service"

	transfers "secure-payment-service/internal/transfers"
)

const (
	fromAccount               = "acc-001"
	toAccount                 = "acc-002"
	amount                    = 100.50
	currency                  = "USD"
	expectedMonitorTransferID = "test-transfer-id-123"
	statusCompleted           = "COMPLETED"
	statusFailed              = "FAILED"
	transferID                = "some-transfer-id"
	expectedBalance           = 123.45
)

func givenAnTransferRequest() transfers.TransferRequest {
	return transfers.TransferRequest{
		FromAccount: fromAccount,
		ToAccount:   toAccount,
		Amount:      amount,
		Currency:    currency,
	}
}

func TestTransferServiceImpl_CreateTransfer_Success(t *testing.T) {
	mockRepo := service.NewMockTransferRepository(t)
	transferService := service.NewTransferService(mockRepo)
	req := givenAnTransferRequest()

	mockRepo.On("CreateTransfer", mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("float64")).
		Return(expectedMonitorTransferID, nil).Once()

	mockRepo.On("GetTransfer", expectedMonitorTransferID).Return(models.Transfer{Status: statusCompleted}, nil).Maybe()

	id, err := transferService.CreateTransfer(req)

	assert.NoError(t, err)
	assert.Equal(t, expectedMonitorTransferID, id)

	time.Sleep(50 * time.Millisecond)
	mockRepo.AssertCalled(t, "CreateTransfer", fromAccount, toAccount, mock.Anything)

	mockRepo.AssertExpectations(t)
}

func TestTransferServiceImpl_CreateTransfer_RepositoryError(t *testing.T) {
	mockRepo := service.NewMockTransferRepository(t)
	expectedError := errors.New("error de base de datos simulado")

	mockRepo.On("CreateTransfer", mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("float64")).
		Return("", expectedError).Once()

	transferService := service.NewTransferService(mockRepo)

	req := givenAnTransferRequest()

	id, err := transferService.CreateTransfer(req)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Empty(t, id)
	mockRepo.AssertExpectations(t)
}

func TestTransferServiceImpl_GetTransfer_Success(t *testing.T) {
	mockRepo := service.NewMockTransferRepository(t)
	transferService := service.NewTransferService(mockRepo)

	expectedTransfer := models.Transfer{
		TransferID:  transferID,
		FromAccount: fromAccount,
		ToAccount:   toAccount,
		Amount:      amount,
		Currency:    currency,
		Status:      statusCompleted,
	}

	mockRepo.On("GetTransfer", transferID).Return(expectedTransfer, nil).Once()

	actualTransfer, err := transferService.GetTransfer(transferID)

	assert.NoError(t, err)
	assert.Equal(t, expectedTransfer, actualTransfer)
	mockRepo.AssertExpectations(t)
	mockRepo.AssertCalled(t, "GetTransfer", transferID)
}

func TestTransferServiceImpl_GetTransfer_RepositoryError(t *testing.T) {
	mockRepo := service.NewMockTransferRepository(t)
	transferService := service.NewTransferService(mockRepo)
	expectedError := errors.New("transfer not found in db")

	mockRepo.On("GetTransfer", "non-existent-id").Return(models.Transfer{}, expectedError).Once()

	actualTransfer, err := transferService.GetTransfer("non-existent-id")

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Equal(t, models.Transfer{}, actualTransfer)
	mockRepo.AssertExpectations(t)
}

func TestTransferServiceImpl_GetAccountBalance_Success(t *testing.T) {
	mockRepo := service.NewMockTransferRepository(t)
	transferService := service.NewTransferService(mockRepo)

	mockRepo.On("GetAccountBalance", toAccount).Return(expectedBalance, nil).Once()

	balance, err := transferService.GetAccountBalance(toAccount)

	assert.NoError(t, err)
	assert.Equal(t, expectedBalance, balance)
	mockRepo.AssertExpectations(t)
}

func TestTransferServiceImpl_GetAccountBalance_RepositoryError(t *testing.T) {
	mockRepo := service.NewMockTransferRepository(t)
	transferService := service.NewTransferService(mockRepo)
	expectedError := errors.New("error fetching balance from repository")

	mockRepo.On("GetAccountBalance", "account-id-error").Return(0.0, expectedError).Once()

	balance, err := transferService.GetAccountBalance("account-id-error")

	assert.Error(t, err)
	assert.Equal(t, 0.0, balance)
	assert.Equal(t, expectedError, err)
	mockRepo.AssertExpectations(t)
}

func TestTransferServiceImpl_UpdateTransfer_Success(t *testing.T) {
	mockRepo := new(service.MockTransferRepository)
	transferService := service.NewTransferService(mockRepo)

	mockRepo.On("UpdateTransfer", transferID, statusCompleted).Return(nil).Once()
	err := transferService.UpdateTransfer(transferID, statusCompleted)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestTransferServiceImpl_UpdateTransfer_RepositoryError(t *testing.T) {
	mockRepo := new(service.MockTransferRepository)
	transferService := service.NewTransferService(mockRepo)
	expectedError := errors.New("error updating transfer in repository")

	mockRepo.On("UpdateTransfer", "transfer-id-error", statusFailed).Return(expectedError).Once()

	err := transferService.UpdateTransfer("transfer-id-error", statusFailed)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockRepo.AssertExpectations(t)
}
