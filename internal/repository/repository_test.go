package repository_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"secure-payment-service/internal/enums"
	"secure-payment-service/internal/models"
	"secure-payment-service/internal/repository"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared&_journal=MEMORY"), &gorm.Config{})
	assert.NoError(t, err, "Fallo al abrir la conexión a SQLite en memoria")

	err = db.AutoMigrate(&models.Transfer{})
	assert.NoError(t, err, "Fallo al auto-migrar el esquema de la base de datos")

	t.Cleanup(func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	})
	return db
}

func TestGormRepository(t *testing.T) {
	mainDB := setupTestDB(t)

	t.Run("CreateTransfer", func(t *testing.T) {
		tx := mainDB.Begin()
		assert.NoError(t, tx.Error)
		defer tx.Rollback()

		repo := repository.NewGormRepository(tx)

		t.Run("success_creation", func(t *testing.T) {
			transferID, err := repo.CreateTransfer("acc_test_from_1", "acc_test_to_1", 150.0)
			assert.NoError(t, err)
			assert.NotEmpty(t, transferID)

			var transfer models.Transfer
			result := tx.Where("transfer_id = ?", transferID).First(&transfer)
			assert.NoError(t, result.Error)
			assert.Equal(t, transferID, transfer.TransferID)
			assert.Equal(t, "acc_test_from_1", transfer.FromAccount)
			assert.Equal(t, "acc_test_to_1", transfer.ToAccount)
			assert.Equal(t, 150.0, transfer.Amount)
			assert.Equal(t, enums.PENDING.String(), transfer.Status)
		})
	})

	t.Run("GetTransfer", func(t *testing.T) {
		tx := mainDB.Begin()
		assert.NoError(t, tx.Error)
		defer tx.Rollback()

		repo := repository.NewGormRepository(tx)

		t.Run("success_found", func(t *testing.T) {
			expectedTransfer := models.Transfer{
				TransferID:  "transfer-xyz-123",
				FromAccount: "sender_acc",
				ToAccount:   "receiver_acc",
				Amount:      200.0,
				Status:      enums.COMPLETED.String(),
			}
			tx.Create(&expectedTransfer)

			foundTransfer, err := repo.GetTransfer("transfer-xyz-123")
			assert.NoError(t, err)
			assert.Equal(t, expectedTransfer.TransferID, foundTransfer.TransferID)
			assert.Equal(t, expectedTransfer.FromAccount, foundTransfer.FromAccount)
			assert.Equal(t, expectedTransfer.ToAccount, foundTransfer.ToAccount)
			assert.Equal(t, expectedTransfer.Amount, foundTransfer.Amount)
			assert.Equal(t, expectedTransfer.Status, foundTransfer.Status)
		})

		t.Run("not_found", func(t *testing.T) {
			_, err := repo.GetTransfer("non-existent-transfer-id")
			assert.Error(t, err)
			assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
		})
	})

	t.Run("GetAccountBalance", func(t *testing.T) {
		tx := mainDB.Begin()
		assert.NoError(t, tx.Error)
		defer tx.Rollback()

		repo := repository.NewGormRepository(tx)

		insertTransfer := func(id, from, to string, amount float64, status string) {
			tx.Create(&models.Transfer{
				TransferID:  id,
				FromAccount: from,
				ToAccount:   to,
				Amount:      amount,
				Status:      status,
			})
		}

		t.Run("non_existent_account_returns_not_found_error", func(t *testing.T) {
			balance, err := repo.GetAccountBalance("acc-empty-001")
			assert.Error(t, err)
			assert.Equal(t, "account not found", err.Error())
			assert.Equal(t, 0.0, balance)
		})

		t.Run("balance_with_completed_transactions", func(t *testing.T) {
			insertTransfer("bal-t1", "other_1", "my_acc", 100.0, enums.COMPLETED.String())
			insertTransfer("bal-t2", "other_2", "my_acc", 50.0, enums.COMPLETED.String())
			insertTransfer("bal-t3", "my_acc", "other_3", 30.0, enums.COMPLETED.String())

			balance, err := repo.GetAccountBalance("my_acc")
			assert.NoError(t, err)
			assert.Equal(t, 120.0, balance)
		})

		t.Run("ignore_pending_and_failed_transactions", func(t *testing.T) {
			insertTransfer("bal-t4", "other_4", "my_acc_2", 200.0, enums.COMPLETED.String())
			insertTransfer("bal-t5", "other_5", "my_acc_2", 70.0, enums.PENDING.String())
			insertTransfer("bal-t6", "my_acc_2", "other_6", 40.0, enums.COMPLETED.String())
			insertTransfer("bal-t7", "my_acc_2", "other_7", 10.0, enums.FAILED.String())

			balance, err := repo.GetAccountBalance("my_acc_2")
			assert.NoError(t, err)
			assert.Equal(t, 160.0, balance)
		})

		t.Run("only_inflows", func(t *testing.T) {
			insertTransfer("bal-t8", "in_src_1", "acc_inonly", 100.0, enums.COMPLETED.String())
			insertTransfer("bal-t9", "in_src_2", "acc_inonly", 50.0, enums.COMPLETED.String())
			balance, err := repo.GetAccountBalance("acc_inonly")
			assert.NoError(t, err)
			assert.Equal(t, 150.0, balance)
		})

		t.Run("only_outflows", func(t *testing.T) {
			insertTransfer("bal-t10", "acc_outonly", "out_dest_1", 70.0, enums.COMPLETED.String())
			insertTransfer("bal-t11", "acc_outonly", "out_dest_2", 30.0, enums.COMPLETED.String())
			balance, err := repo.GetAccountBalance("acc_outonly")
			assert.NoError(t, err)
			assert.Equal(t, -100.0, balance)
		})

		t.Run("null_float_handling", func(t *testing.T) {
			insertTransfer("nf-t1", "acc-nf", "dest", 100, enums.COMPLETED.String())
			balance, err := repo.GetAccountBalance("acc-nf")
			assert.NoError(t, err)
			assert.Equal(t, -100.0, balance)

			insertTransfer("nf-t2", "src", "acc-nf-2", 50, enums.COMPLETED.String())
			balance2, err := repo.GetAccountBalance("acc-nf-2")
			assert.NoError(t, err)
			assert.Equal(t, 50.0, balance2)
		})
	})

	t.Run("UpdateTransfer", func(t *testing.T) {
		tx := mainDB.Begin()
		assert.NoError(t, tx.Error)
		defer tx.Rollback()

		repo := repository.NewGormRepository(tx)

		t.Run("success_update_status", func(t *testing.T) {
			initialTransfer := models.Transfer{
				TransferID:  "update-id-456",
				FromAccount: "userA",
				ToAccount:   "userB",
				Amount:      100.0,
				Status:      "PENDING",
			}
			tx.Create(&initialTransfer)

			err := repo.UpdateTransfer("update-id-456", enums.COMPLETED.String())
			assert.NoError(t, err)

			var updatedTransfer models.Transfer
			result := tx.Where("transfer_id = ?", "update-id-456").First(&updatedTransfer)
			assert.NoError(t, result.Error)
			assert.Equal(t, enums.COMPLETED.String(), updatedTransfer.Status)
		})

		t.Run("transfer_not_found", func(t *testing.T) {
			err := repo.UpdateTransfer("non-existent-update-id", enums.COMPLETED.String())
			assert.Error(t, err)
			assert.EqualError(t, err, "transfer not found")
		})

		t.Run("no_change_if_status_is_same", func(t *testing.T) {
			initialTransfer := models.Transfer{
				TransferID:  "update-id-789",
				FromAccount: "userX",
				ToAccount:   "userY",
				Amount:      50.0,
				Status:      enums.PENDING.String(),
			}
			tx.Create(&initialTransfer)

			err := repo.UpdateTransfer("update-id-789", enums.PENDING.String())
			assert.NoError(t, err)
		})
	})
}
