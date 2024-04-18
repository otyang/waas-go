package account

import (
	"context"
	"testing"
	"time"

	"github.com/otyang/waas-go/types"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestClient_Transaction_and_all_its_assosiated_functions(t *testing.T) {
	t.Parallel()

	a := &Client{db: TestDB}

	txn := types.Transaction{
		ID:         "txn_12345678",               // Generate a unique transaction ID.
		CustomerID: "cust_9876543",               // Some customer identifier
		WalletID:   "wallet_54321",               // Some wallet identifier
		IsDebit:    true,                         // Credit transaction
		Currency:   "USD",                        // Or any other relevant currency code (e.g., "NGN")
		Amount:     decimal.NewFromFloat(100.50), // Amount of the transaction
		Fee:        decimal.NewFromFloat(2.50),   // Transaction fee
		//	TotalAmount:    decimal.NewFromFloat(102.50),
		BalanceAfter:   decimal.NewFromFloat(250.25), // Assuming a previous balance
		Type:           types.TransactionTypeSwap,
		Status:         types.TransactionStatusNew,
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
		got, err := a.UpdateTransactionStatus(context.Background(), txn.ID, types.TransactionStatusFailed)
		assert.NoError(t, err)
		assert.NotEqual(t, "wont_update_this_12345", got.WalletID)
		assert.Equal(t, types.TransactionStatusFailed, got.Status)
	})

	t.Run("update transaction status", func(t *testing.T) {
		txn.WalletID = "new_cust_id_12345"
		got, err := a.UpdateTransaction(context.Background(), &txn)
		assert.NoError(t, err)
		assert.Equal(t, txn.WalletID, got.WalletID)
	})

	t.Run("list without filters", func(t *testing.T) {
		gotList, nextCursor, err := a.ListTransaction(context.Background(), types.ListTransactionsFilterOpts{
			Limit: 400,
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, nextCursor)

		//	t.Errorf("%+v", gotList)
		assert.NotEmpty(t, gotList)
	})

	t.Run("list with filters", func(t *testing.T) {
		gotList, nextCursor, err := a.ListTransaction(context.Background(), types.ListTransactionsFilterOpts{
			Limit:      0,
			StartDate:  time.Time{},
			EndDate:    time.Time{},
			CustomerID: "",
			WalletID:   "",
			Currency:   []string{"usD"},
			IsDebit:    toPointer(true),
			Type:       toPointer(types.TransactionTypeSwap),
			Status:     toPointer(types.TransactionStatusNew),
			Reversed:   toPointer(false),
		})
		assert.NoError(t, err)
		assert.Empty(t, nextCursor)
		assert.NotEmpty(t, gotList)
	})
}
