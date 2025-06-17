package enums_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"secure-payment-service/internal/enums"
)

func TestTransactionStatus_String(t *testing.T) {
	tests := []struct {
		status   enums.TransactionStatus
		expected string
	}{
		{enums.COMPLETED, "COMPLETED"},
		{enums.PENDING, "PENDING"},
		{enums.FAILED, "FAILED"},
		{"UNKNOWN_STATUS", "UNKNOWN_STATUS"},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.status.String())
		})
	}
}

func TestTransactionStatus_IsValid(t *testing.T) {
	tests := []struct {
		status   enums.TransactionStatus
		expected bool
	}{
		{enums.COMPLETED, true},
		{enums.PENDING, true},
		{enums.FAILED, true},
		{"", false},
		{"completed", false},
		{"COMPLEETED", false},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.status.IsValid())
		})
	}
}

func TestNewTransactionStatusFromString(t *testing.T) {
	tests := []struct {
		input    string
		expected enums.TransactionStatus
		err      bool
	}{
		{"COMPLETED", enums.COMPLETED, false},
		{"PENDING", enums.PENDING, false},
		{"FAILED", enums.FAILED, false},
		{"INVALID", "", true},
		{"", "", true},
		{"pending", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			status, err := enums.NewTransactionStatusFromString(tt.input)
			if tt.err {
				assert.Error(t, err)
				assert.Empty(t, status)
				assert.Contains(t, err.Error(), "is not a valid transaction status")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, status)
			}
		})
	}
}
