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
	timer := prometheus.NewTimer(metrics.ServiceOperationDurationSeconds.WithLabelValues("create_transfer", "success"))
	defer timer.ObserveDuration()

	id, err := s.repo.CreateTransfer(req.FromAccount, req.ToAccount, req.Amount)
	if err != nil {
		metrics.ServiceOperationsTotal.WithLabelValues("create_transfer", "failure").Inc()
		timer.ObserveDuration()
		metrics.ServiceOperationDurationSeconds.WithLabelValues("create_transfer", "failure").Observe(0)
		return "", err
	}

	metrics.ServiceOperationsTotal.WithLabelValues("create_transfer", "success").Inc()
	go s.MonitorTransfer(id)

	return id, nil
}

func (s *TransferServiceImpl) GetTransfer(id string) (models.Transfer, error) {
	timer := prometheus.NewTimer(metrics.ServiceOperationDurationSeconds.WithLabelValues("get_transfer", "success"))
	defer timer.ObserveDuration()

	transfer, err := s.repo.GetTransfer(id)
	if err != nil {
		statusLabel := "failure"
		if errors.Is(err, errors.New("transfer not found")) {
			statusLabel = "not_found"
		}
		metrics.ServiceOperationsTotal.WithLabelValues("get_transfer", statusLabel).Inc()
		timer.ObserveDuration()
		metrics.ServiceOperationDurationSeconds.WithLabelValues("get_transfer", statusLabel).Observe(0)
		return models.Transfer{}, err
	}

	metrics.ServiceOperationsTotal.WithLabelValues("get_transfer", "success").Inc()
	return transfer, nil
}

func (s *TransferServiceImpl) GetAccountBalance(id string) (float64, error) {
	timer := prometheus.NewTimer(metrics.ServiceOperationDurationSeconds.WithLabelValues("get_account_balance", "success"))
	defer timer.ObserveDuration()

	balance, err := s.repo.GetAccountBalance(id)
	if err != nil {
		statusLabel := "failure"
		if errors.Is(err, errors.New("account not found")) {
			statusLabel = "not_found"
		}
		metrics.ServiceOperationsTotal.WithLabelValues("get_account_balance", statusLabel).Inc()
		timer.ObserveDuration()
		metrics.ServiceOperationDurationSeconds.WithLabelValues("get_account_balance", statusLabel).Observe(0)
		return 0, err
	}

	metrics.ServiceOperationsTotal.WithLabelValues("get_account_balance", "success").Inc()
	return balance, nil
}

func (s *TransferServiceImpl) UpdateTransfer(id, status string) error {
	timer := prometheus.NewTimer(metrics.ServiceOperationDurationSeconds.WithLabelValues("update_transfer", "success"))
	defer timer.ObserveDuration()

	if err := s.repo.UpdateTransfer(id, status); err != nil {
		statusLabel := "failure"
		if errors.Is(err, errors.New("transfer not found")) {
			statusLabel = "not_found"
		}
		metrics.ServiceOperationsTotal.WithLabelValues("update_transfer", statusLabel).Inc()
		timer.ObserveDuration()
		metrics.ServiceOperationDurationSeconds.WithLabelValues("update_transfer", statusLabel).Observe(0)
		return err
	}

	metrics.ServiceOperationsTotal.WithLabelValues("update_transfer", "success").Inc()
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
			metrics.TransferMonitorAttemptsTotal.WithLabelValues(id, fmt.Sprintf("%d", i+1), "success").Inc()
			return
		}
	}
	metrics.TransferMonitorAttemptsTotal.WithLabelValues(id, fmt.Sprintf("%d", maxAttempts), "max_attempts_reached").Inc()
}
