package service

import (
	"errors"
	"fmt"
	"time"

	"secure-payment-service/internal/enums"
	"secure-payment-service/internal/metrics"
	"secure-payment-service/internal/models"
	"secure-payment-service/internal/repository"
	"secure-payment-service/internal/transfers"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	StatusSuccess     = "success"
	StatusFailure     = "failure"
	StatusNotFound    = "not_found"
	CreateTransfer    = "create_transfer"
	GetTransfer       = "get_transfer"
	GetAccountBalance = "get_account_balance"
	UpdateTransfer    = "update_transfer"
)

type TransferService interface {
	CreateTransfer(req transfers.TransferRequest) (string, error)
	GetTransfer(id string) (models.Transfer, error)
	GetAccountBalance(id string) (float64, error)
	UpdateTransfer(id, status string) error
}

type TransferServiceImpl struct {
	repo repository.TransferRepository
}

func NewTransferService(repo repository.TransferRepository) TransferService {
	return &TransferServiceImpl{repo: repo}
}

func (s *TransferServiceImpl) CreateTransfer(req transfers.TransferRequest) (string, error) {
	timer := prometheus.NewTimer(metrics.ServiceOperationDurationSeconds.WithLabelValues(CreateTransfer, StatusSuccess))
	defer timer.ObserveDuration()

	id, err := s.repo.CreateTransfer(req.FromAccount, req.ToAccount, req.Amount)
	if err != nil {
		metrics.ServiceOperationsTotal.WithLabelValues(CreateTransfer, StatusFailure).Inc()
		timer.ObserveDuration()
		metrics.ServiceOperationDurationSeconds.WithLabelValues(CreateTransfer, StatusFailure).Observe(0)
		return "", err
	}

	metrics.ServiceOperationsTotal.WithLabelValues(CreateTransfer, StatusSuccess).Inc()
	go s.MonitorTransfer(id)

	return id, nil
}

func (s *TransferServiceImpl) GetTransfer(id string) (models.Transfer, error) {
	timer := prometheus.NewTimer(metrics.ServiceOperationDurationSeconds.WithLabelValues(GetTransfer, StatusSuccess))
	defer timer.ObserveDuration()

	transfer, err := s.repo.GetTransfer(id)
	if err != nil {
		statusLabel := StatusFailure
		if errors.Is(err, errors.New("transfer not found")) {
			statusLabel = StatusNotFound
		}
		metrics.ServiceOperationsTotal.WithLabelValues(GetTransfer, statusLabel).Inc()
		timer.ObserveDuration()
		metrics.ServiceOperationDurationSeconds.WithLabelValues(GetTransfer, statusLabel).Observe(0)
		return models.Transfer{}, err
	}

	metrics.ServiceOperationsTotal.WithLabelValues(GetTransfer, StatusSuccess).Inc()
	return transfer, nil
}

func (s *TransferServiceImpl) GetAccountBalance(id string) (float64, error) {
	timer := prometheus.NewTimer(metrics.ServiceOperationDurationSeconds.WithLabelValues(GetAccountBalance, StatusSuccess))
	defer timer.ObserveDuration()

	balance, err := s.repo.GetAccountBalance(id)
	if err != nil {
		statusLabel := StatusFailure
		if errors.Is(err, errors.New("account not found")) {
			statusLabel = StatusNotFound
		}
		metrics.ServiceOperationsTotal.WithLabelValues(GetAccountBalance, statusLabel).Inc()
		timer.ObserveDuration()
		metrics.ServiceOperationDurationSeconds.WithLabelValues(GetAccountBalance, statusLabel).Observe(0)
		return 0, err
	}

	metrics.ServiceOperationsTotal.WithLabelValues(GetAccountBalance, StatusSuccess).Inc()
	return balance, nil
}

func (s *TransferServiceImpl) UpdateTransfer(id, status string) error {
	timer := prometheus.NewTimer(metrics.ServiceOperationDurationSeconds.WithLabelValues(UpdateTransfer, StatusSuccess))
	defer timer.ObserveDuration()

	if err := s.repo.UpdateTransfer(id, status); err != nil {
		statusLabel := StatusFailure
		if errors.Is(err, errors.New("transfer not found")) {
			statusLabel = StatusNotFound
		}
		metrics.ServiceOperationsTotal.WithLabelValues(UpdateTransfer, statusLabel).Inc()
		timer.ObserveDuration()
		metrics.ServiceOperationDurationSeconds.WithLabelValues(UpdateTransfer, statusLabel).Observe(0)
		return err
	}

	metrics.ServiceOperationsTotal.WithLabelValues(UpdateTransfer, StatusSuccess).Inc()
	return nil
}

func (s *TransferServiceImpl) MonitorTransfer(id string) {
	const maxAttempts = 5
	const baseDelay = 5

	for i := 0; i < maxAttempts; i++ {
		transfer, err := s.GetTransfer(id)
		if err == nil && transfer.Status == enums.PENDING.String() {
			metrics.TransferMonitorAttemptsTotal.WithLabelValues(id, fmt.Sprintf("%d", i+1), "still_pending").Inc()
			time.Sleep(time.Duration(baseDelay+i*2) * time.Second)
			continue
		} else if err != nil {
			metrics.TransferMonitorAttemptsTotal.WithLabelValues(id, fmt.Sprintf("%d", i+1), "error").Inc()
			return
		} else {
			metrics.TransferMonitorAttemptsTotal.WithLabelValues(id, fmt.Sprintf("%d", i+1), StatusSuccess).Inc()
			return
		}
	}
	metrics.TransferMonitorAttemptsTotal.WithLabelValues(id, fmt.Sprintf("%d", maxAttempts), "max_attempts_reached").Inc()
}
