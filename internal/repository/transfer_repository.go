package repository

import (
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"secure-payment-service/internal/enums"
	"secure-payment-service/internal/models"
)

type TransferRepository interface {
	CreateTransfer(from, to string, amount float64) (string, error)
	GetTransfer(id string) (models.Transfer, error)
	GetAccountBalance(id string) (float64, error)
	UpdateTransfer(id, status string) error
}

type GormRepository struct {
	db *gorm.DB
}

func NewGormRepository(database *gorm.DB) TransferRepository {
	return &GormRepository{db: database}
}

func generateUUID() string {
	return uuid.New().String()
}

func (r *GormRepository) CreateTransfer(from, to string, amount float64) (string, error) {
	transfer := models.Transfer{
		TransferID:  generateUUID(),
		FromAccount: from,
		ToAccount:   to,
		Amount:      amount,
		Status:      enums.PENDING.String(),
	}

	if err := r.db.Create(&transfer).Error; err != nil {
		return "", err
	}

	return transfer.TransferID, nil
}

func (r *GormRepository) GetTransfer(id string) (models.Transfer, error) {
	var transfer models.Transfer
	result := r.db.Where("transfer_id = ?", id).First(&transfer)
	if result.Error != nil {
		return transfer, result.Error
	}

	return transfer, nil
}

func (r *GormRepository) GetAccountBalance(id string) (float64, error) {
	var balanceIn sql.NullFloat64
	var balanceOut sql.NullFloat64

	resultInErr := r.db.Model(&models.Transfer{}).
		Select("sum(amount)").
		Where("to_account = ? AND status = ?", id, enums.COMPLETED.String()).
		Row().Scan(&balanceIn)

	if resultInErr != nil && resultInErr != sql.ErrNoRows {
		return 0, resultInErr
	}

	resultOutErr := r.db.Model(&models.Transfer{}).
		Select("sum(amount)").
		Where("from_account = ? AND status = ?", id, enums.COMPLETED.String()).
		Row().Scan(&balanceOut)

	if resultOutErr != nil && resultOutErr != sql.ErrNoRows {
		return 0, resultOutErr
	}

	calculatedBalance := balanceIn.Float64 - balanceOut.Float64

	return calculatedBalance, nil
}

func (r *GormRepository) UpdateTransfer(id, status string) error {
	result := r.db.Model(&models.Transfer{}).Where("transfer_id = ?", id).Update("status", status)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("transfer not found")
	}

	return nil
}
