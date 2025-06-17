package enums

import "fmt"

type TransactionStatus string

const (
	COMPLETED TransactionStatus = "COMPLETED"
	PENDING   TransactionStatus = "PENDING"
	FAILED    TransactionStatus = "FAILED"
)

func (ts TransactionStatus) String() string {
	return string(ts)
}

func (ts TransactionStatus) IsValid() bool {
	switch ts {
	case COMPLETED, PENDING, FAILED:
		return true
	default:
		return false
	}
}

func NewTransactionStatusFromString(s string) (TransactionStatus, error) {
	status := TransactionStatus(s)
	if !status.IsValid() {
		return "", fmt.Errorf("'%s' is not a valid transaction status", s)
	}
	return status, nil
}
