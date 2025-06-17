package service

import (
	"time"

	"secure-payment-service/internal/enums"
	"secure-payment-service/internal/models"
	"secure-payment-service/internal/repository"
	"secure-payment-service/internal/transfers"
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
	id, err := s.repo.CreateTransfer(req.FromAccount, req.ToAccount, req.Amount)
	if err != nil {
		return "", err
	}

	go s.MonitorTransfer(id)

	return id, nil
}

func (s *TransferServiceImpl) GetTransfer(id string) (models.Transfer, error) {
	transfer, err := s.repo.GetTransfer(id)
	if err != nil {
		return models.Transfer{}, err
	}

	return transfer, nil
}

func (s *TransferServiceImpl) GetAccountBalance(id string) (float64, error) {
	balance, err := s.repo.GetAccountBalance(id)
	if err != nil {
		return 0, err
	}

	return balance, nil
}

func (s *TransferServiceImpl) UpdateTransfer(id, status string) error {
	if err := s.repo.UpdateTransfer(id, status); err != nil {
		return err
	}

	return nil
}

func (s *TransferServiceImpl) MonitorTransfer(id string) {
	const maxAttempts = 5
	const baseDelay = 5

	for i := 0; i < maxAttempts; i++ {
		transfer, err := s.GetTransfer(id)
		if err == nil && transfer.Status == enums.PENDING.String() {
			time.Sleep(time.Duration(baseDelay+i*2) * time.Second)
			continue
		}
		return
	}
}
