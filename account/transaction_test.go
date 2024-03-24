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
		ID:         "txn_12345678",               // Generate a unique transaction ID
		CustomerID: "cust_9876543",               // Some customer identifier
		WalletID:   "wallet_54321",               // Some wallet identifier
		IsDebit:    true,                         // Credit transaction
		Currency:   "USD",                        // Or any other relevant currency code (e.g., "NGN")
		Amount:     decimal.NewFromFloat(100.50), // Amount of the transaction
		Fee:        decimal.NewFromFloat(2.50),   // Transaction fee
		//	TotalAmount:    decimal.NewFromFloat(102.50),
		BalanceAfter:   decimal.NewFromFloat(250.25), // Assuming a previous balance
		Type:           waas.TransactionTypeSwap,
		Status:         waas.TransactionStatusNew,
		Narration:      toPointer("Payment for Order ABC"),
		CounterpartyID: toPointer("tx_123456"),
		CreatedAt:      time.Time{},
		UpdatedAt:      time.Time{},
	}

	_, err := a.CreateTransaction(context.Background(), &txn)
	assert.NoError(t, err)

	t.Run("get transaction", func(t *testing.T) {
		got, err := a.GetTransaction(context.Background(), txn.ID)
		assert.NoError(t, err)
		assert.Equal(t, txn.ID, got.ID)
	})

	t.Run("update transaction", func(t *testing.T) {
		txn.WalletID = "wont_update_this_12345"
		got, err := a.UpdateTransactionStatus(context.Background(), txn.ID, waas.TransactionStatusFailed)
		assert.NoError(t, err)
		assert.NotEqual(t, "wont_update_this_12345", got.WalletID)
		assert.Equal(t, waas.TransactionStatusFailed, got.Status)
	})

	t.Run("update transaction status", func(t *testing.T) {
		txn.WalletID = "new_cust_id_12345"
		got, err := a.UpdateTransaction(context.Background(), &txn)
		assert.NoError(t, err)
		assert.Equal(t, txn.WalletID, got.WalletID)
	})

	t.Run("list without filters", func(t *testing.T) {
		gotList, nextCursor, err := a.ListTransaction(context.Background(), waas.ListTransactionsFilterParams{})
		assert.NoError(t, err)
		assert.Empty(t, nextCursor)
		assert.NotEmpty(t, gotList)
	})

	t.Run("list with filters", func(t *testing.T) {
		gotList, nextCursor, err := a.ListTransaction(context.Background(), waas.ListTransactionsFilterParams{
			Limit:      0,
			Before:     time.Time{},
			After:      time.Time{},
			CustomerID: nil,
			WalletID:   nil,
			Currency:   []string{"usD"},
			IsDebit:    toPointer(true),
			Type:       toPointer(waas.TransactionTypeSwap),
			Status:     toPointer(waas.TransactionStatusNew),
			Reversed:   toPointer(false),
		})
		assert.NoError(t, err)
		assert.Empty(t, nextCursor)
		assert.NotEmpty(t, gotList)
	})
}
