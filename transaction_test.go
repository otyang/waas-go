package waas

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestTransaction_Reverse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		tx        *Transaction
		wallet    *Wallet
		wantErr   bool
		wantError error
	}{
		{
			name:      "Reverse successful withdrawal",
			tx:        &Transaction{Type: TransactionTypeWithdrawal, Status: TransactionStatusSuccess, IsDebit: true, Amount: decimal.NewFromFloat(100), Fee: decimal.NewFromFloat(5)},
			wallet:    &Wallet{AvailableBalance: decimal.NewFromFloat(500)},
			wantErr:   false,
			wantError: nil,
		},
		{
			name:      "Reverse successful deposit",
			tx:        &Transaction{Type: TransactionTypeDeposit, Status: TransactionStatusSuccess, IsDebit: false, Amount: decimal.NewFromFloat(100), Fee: decimal.NewFromFloat(5)},
			wallet:    &Wallet{AvailableBalance: decimal.NewFromFloat(500)},
			wantErr:   true,
			wantError: ErrUnsupportedReversalType,
		},
		{
			name:      "Error on reversing non-withdrawal transaction",
			tx:        &Transaction{Type: TransactionTypeDeposit, Status: TransactionStatusSuccess, IsDebit: false, Amount: decimal.NewFromFloat(100), Fee: decimal.NewFromFloat(5)},
			wallet:    &Wallet{},
			wantErr:   true,
			wantError: ErrUnsupportedReversalType,
		},
		{
			name:      "Error on reversing settled transaction",
			tx:        &Transaction{Type: TransactionTypeWithdrawal, Status: TransactionStatusSuccess, IsDebit: true, Amount: decimal.NewFromFloat(100), Fee: decimal.NewFromFloat(5), Reversed: true},
			wallet:    &Wallet{},
			wantErr:   true,
			wantError: ErrAlreadyReversedTx,
		},
		{
			name:      "Error on reversing nil transaction",
			tx:        nil,
			wallet:    &Wallet{},
			wantErr:   true,
			wantError: ErrInvalidTransactionObject,
		},
		// Add more test cases for different scenarios
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.tx.Reverse(tt.wallet)

			assert.Equal(t, tt.wantError, err)
			assert.Equal(t, tt.wantErr, (err != nil))

			if tt.wantErr {
				assert.Error(t, tt.wantError)
			}
		})
	}
}
