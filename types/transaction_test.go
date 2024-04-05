package types

import (
	"strings"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestTransaction_SetNarrationAndCounterPartyID(t *testing.T) {
	hWorldPtr := strings.TrimSpace("Hello World")

	testCases := []struct {
		name     string
		input    string
		expected *string // Expect a pointer to a string
	}{
		{"empty string", "", nil},
		{"empty space string", "  ", nil},
		{"trims whitespace", "   Hello World   ", &hWorldPtr},
	}

	// Test logic
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			transaction := &Transaction{}
			transaction.SetNarration(tc.input)
			transaction.SetCounterpartyID(tc.input)

			if (tc.expected == nil && transaction.Narration != nil) ||
				(tc.expected != nil && *transaction.Narration != *tc.expected) {
				t.Errorf("Expected narration %v, got %v", tc.expected, transaction.Narration)
			}

			if (tc.expected == nil && transaction.CounterpartyID != nil) ||
				(tc.expected != nil && *transaction.CounterpartyID != *tc.expected) {
				t.Errorf("Expected counterparty %v, got %v", tc.expected, transaction.CounterpartyID)
			}
		})
	}
}

func TestTransaction_canBeReversed(t *testing.T) {
	timePtr := time.Now()

	// Test Cases
	testCases := []struct {
		name          string
		transaction   *Transaction
		expectedError error
	}{
		{
			name: "Valid Reversal",
			transaction: &Transaction{
				Type:   TransactionTypeWithdrawal,
				Status: TransactionStatusSuccess,
			},
			expectedError: nil,
		},
		{
			name:          "Nil Transaction",
			transaction:   nil,
			expectedError: ErrInvalidTransactionObject,
		},
		{
			name: "Invalid Type (Deposit)",
			transaction: &Transaction{
				Type: TransactionTypeDeposit, // Not a withdrawal
			},
			expectedError: ErrTransactionUnsupportedReversalType,
		},
		{
			name: "Already Reversed",
			transaction: &Transaction{
				Type:       TransactionTypeWithdrawal,
				ReversedAt: &timePtr,
			},
			expectedError: ErrTransactionAlreadyReversed,
		},
	}

	// Run Test Cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.transaction.canBeReversed()
			assert.Equal(t, tc.expectedError, err)
		})
	}
}

func TestTransaction_Reverse(t *testing.T) {
	t.Parallel()

	timePtr := time.Now()

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
			wantError: ErrTransactionUnsupportedReversalType,
		},
		{
			name:      "Error on reversing non-withdrawal transaction",
			tx:        &Transaction{Type: TransactionTypeDeposit, Status: TransactionStatusSuccess, IsDebit: false, Amount: decimal.NewFromFloat(100), Fee: decimal.NewFromFloat(5)},
			wallet:    &Wallet{},
			wantErr:   true,
			wantError: ErrTransactionUnsupportedReversalType,
		},
		{
			name:      "Error on reversing already revered transaction",
			tx:        &Transaction{Type: TransactionTypeWithdrawal, ReversedAt: &timePtr, IsDebit: true, Amount: decimal.NewFromFloat(100), Fee: decimal.NewFromFloat(5)},
			wallet:    &Wallet{},
			wantErr:   true,
			wantError: ErrTransactionAlreadyReversed,
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

func Test_NewTransactionForCreditEntry(t *testing.T) {
	mockWallet := Wallet{
		ID:               "test-wallet-id",
		CustomerID:       "test-customer-id",
		CurrencyCode:     "USD",
		AvailableBalance: decimal.NewFromInt(100),
	}

	amount := decimal.NewFromFloat(25.50)
	fee := decimal.NewFromFloat(0.50)
	txnType := TransactionTypeDeposit // Or any relevant type

	// Call the function
	transaction := NewTransactionForCreditEntry(&mockWallet, amount, fee, txnType)

	// Assertions
	assert.NotNil(t, transaction)
	assert.Equal(t, mockWallet.CustomerID, transaction.CustomerID)
	assert.Equal(t, mockWallet.ID, transaction.WalletID)
	assert.Equal(t, false, transaction.IsDebit)
	assert.Equal(t, amount, transaction.Amount)
	assert.Equal(t, fee, transaction.Fee)
	assert.Equal(t, txnType, transaction.Type)
	assert.Equal(t, TransactionStatusSuccess, transaction.Status)
	assert.Equal(t, mockWallet.AvailableBalance, transaction.BalanceAfter)
	assert.WithinDuration(t, transaction.CreatedAt, time.Now(), 1*time.Second)
	assert.WithinDuration(t, transaction.UpdatedAt, time.Now(), 1*time.Second)
}

func Test_NewTransactionForDebitEntry(t *testing.T) {
	mockWallet := Wallet{
		ID:               "test-wallet-id",
		CustomerID:       "test-customer-id",
		CurrencyCode:     "USD",
		AvailableBalance: decimal.NewFromInt(100),
	}

	amount := decimal.NewFromFloat(25.50)
	fee := decimal.NewFromFloat(0.50)
	txnType := TransactionTypeDeposit // Or any relevant type

	// Call the function
	transaction := NewTransactionForDebitEntry(&mockWallet, amount, fee, txnType, TransactionStatusFailed)

	// Assertions
	assert.NotNil(t, transaction)
	assert.Equal(t, mockWallet.CustomerID, transaction.CustomerID)
	assert.Equal(t, mockWallet.ID, transaction.WalletID)
	assert.Equal(t, true, transaction.IsDebit)
	assert.Equal(t, amount, transaction.Amount)
	assert.Equal(t, fee, transaction.Fee)
	assert.Equal(t, txnType, transaction.Type)
	assert.Equal(t, TransactionStatusFailed, transaction.Status)
	assert.Equal(t, mockWallet.AvailableBalance, transaction.BalanceAfter)
	assert.WithinDuration(t, transaction.CreatedAt, time.Now(), 1*time.Second)
	assert.WithinDuration(t, transaction.UpdatedAt, time.Now(), 1*time.Second)
}
