package account

import (
	"context"
	"testing"
	"time"

	"github.com/otyang/waas-go"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestAccount_Transaction_and_all_its_assosiated_functions(t *testing.T) {
	t.Parallel()

	a := &Account{
		db: TestDB,
	}

	txn := waas.Transaction{
		ID:               "txn_12345678",               // Generate a unique transaction ID
		CustomerID:       "cust_9876543",               // Some customer identifier
		WalletID:         "wallet_54321",               // Some wallet identifier
		IsDebit:          true,                         // Credit transaction
		Currency:         "USD",                        // Or any other relevant currency code (e.g., "NGN")
		Amount:           decimal.NewFromFloat(100.50), // Amount of the transaction
		Fee:              decimal.NewFromFloat(2.50),   // Transaction fee
		TotalAmount:      decimal.NewFromFloat(102.50),
		BalanceAfter:     decimal.NewFromFloat(250.25), // Assuming a previous balance
		OriginalCurrency: "",
		OriginalAmount:   decimal.Decimal{},
		OriginalFee:      decimal.Decimal{},
		Type:             waas.TransactionTypeSwap,
		Status:           waas.TransactionStatusNew,
		Narration:        toPointer("Payment for Order ABC"),
		Reversed:         false,
		CounterpartyID:   toPointer("tx_123456"),
		CreatedAt:        time.Time{},
		UpdatedAt:        time.Time{},
	}

	_, err := a.CreateTransaction(context.Background(), &txn)
	assert.NoError(t, err)

	t.Run("get transaction", func(t *testing.T) {
		got, err := a.GetTransaction(context.Background(), txn.ID)
		assert.NoError(t, err)
		assert.Equal(t, txn.ID, got.ID)
	})

	t.Run("update transaction", func(t *testing.T) {
		txn.WalletID = "new_cust_id_12345"
		got, err := a.UpdateTransaction(context.Background(), &txn)
		assert.NoError(t, err)
		assert.Equal(t, txn.WalletID, got.WalletID)
	})

	t.Run("list without filters", func(t *testing.T) {
		gotList, err := a.ListTransaction(context.Background(), 0, waas.ListTransactionsFilterParams{})
		assert.NoError(t, err)
		assert.NotEmpty(t, gotList)
	})

	t.Run("list with filters", func(t *testing.T) {
		gotList, err := a.ListTransaction(context.Background(), 0, waas.ListTransactionsFilterParams{
			After:      time.Time{},
			Before:     time.Time{},
			CustomerID: nil,
			WalletID:   nil,
			Currency:   toPointer("usD"),
			IsDebit:    toPointer(true),
			Type:       toPointer(waas.TransactionTypeSwap),
			Status:     toPointer(waas.TransactionStatusNew),
			Reversed:   toPointer(false),
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, gotList)
	})
}
