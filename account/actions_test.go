package account

import (
	"context"
	"strings"
	"testing"

	"github.com/otyang/waas-go/types"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestClient_WithTxBulkUpdateWalletAndInsertTransaction(t *testing.T) {
	t.Parallel()

	acc := &Client{
		db: TestDB,
	}

	wallet, err := acc.CreateWallet(context.Background(), createTestRandomWallet("cust_800", "ngn"))
	assert.NoError(t, err)

	got, err := acc.Credit(context.Background(), types.CreditWalletOpts{
		WalletID:  wallet.ID,
		Amount:    decimal.NewFromFloat(50),
		Fee:       decimal.Zero,
		Type:      types.TransactionTypeDeposit,
		Narration: "deposit of funds",
	})
	assert.NoError(t, err)

	err = acc.WithTxBulkUpdateWalletAndInsertTransaction(
		context.Background(), []*types.Wallet{got.Wallet}, []*types.Transaction{got.Transaction},
	)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "constraint"))
}

func TestClient_Credit_and_Debit(t *testing.T) {
	t.Parallel()

	acc := &Client{
		db: TestDB,
	}

	w := createTestRandomWallet("cust_1404", "ngn")

	_, err := acc.CreateWallet(context.Background(), w)
	assert.NoError(t, err)

	t.Run("crediting", func(t *testing.T) {
		got, err := acc.Credit(context.Background(), types.CreditWalletOpts{
			WalletID:  w.ID,
			Amount:    decimal.NewFromFloat(50),
			Fee:       decimal.Zero,
			Type:      types.TransactionTypeDeposit,
			Narration: "deposit of funds",
		})

		assert.NoError(t, err)
		assert.Equal(t, got.Wallet.AvailableBalance.String(), decimal.NewFromFloat(50).String())
		assert.Equal(t, types.TransactionTypeDeposit, got.Transaction.Type)
		assert.Equal(t, "deposit of funds", *got.Transaction.Narration)
	})

	t.Run("debiting", func(t *testing.T) {
		got, err := acc.Debit(context.Background(), types.DebitWalletOpts{
			WalletID:  w.ID,
			Amount:    decimal.NewFromFloat(20),
			Fee:       decimal.Zero,
			Type:      types.TransactionTypeWithdrawal,
			Narration: "withdrawal of funds",
		})

		assert.NoError(t, err)
		assert.Equal(t, got.Wallet.AvailableBalance.String(), decimal.NewFromFloat(30).String())
		assert.Equal(t, types.TransactionTypeWithdrawal, got.Transaction.Type)
		assert.Equal(t, "withdrawal of funds", *got.Transaction.Narration)
	})
}

func TestClient_Transfer(t *testing.T) {
	t.Parallel()

	acc := &Client{
		db: TestDB,
	}

	sourceWallet := createTestRandomWallet("cust_1", "ngn")
	sourceWallet.AvailableBalance = decimal.NewFromFloat(30)
	destinationWallet := createTestRandomWallet("cust_2", "ngn")

	_, err := acc.CreateWallet(context.Background(), sourceWallet)
	assert.NoError(t, err)

	_, err = acc.CreateWallet(context.Background(), destinationWallet)
	assert.NoError(t, err)

	got, err := acc.Transfer(context.Background(), types.TransferRequestOpts{
		FromWalletID: sourceWallet.ID,
		ToWalletID:   destinationWallet.ID,
		Amount:       decimal.NewFromFloat(30),
		Fee:          decimal.Zero,
		Narration:    "transfer of funds",
	})

	assert.NoError(t, err)
	assert.Equal(t, decimal.Zero.String(), got.FromWallet.AvailableBalance.String())
	assert.Equal(t, decimal.NewFromFloat(30).String(), got.ToWallet.AvailableBalance.String())
	assert.Equal(t, types.TransactionTypeTransfer, got.FromTransaction.Type)
	assert.Equal(t, types.TransactionTypeTransfer, got.ToTransaction.Type)
	assert.Equal(t, "transfer of funds", *got.FromTransaction.Narration)
	assert.Equal(t, "transfer of funds", *got.ToTransaction.Narration)
}

func TestClient_Swap(t *testing.T) {
	t.Parallel()

	acc := &Client{
		db: TestDB,
	}

	fromWallet := createTestRandomWallet("cust_10", "ngn")
	fromWallet.AvailableBalance = decimal.NewFromFloat(1500)
	toWallet := createTestRandomWallet("cust_10", "usd")

	_, err := acc.CreateWallet(context.Background(), fromWallet)
	assert.NoError(t, err)

	_, err = acc.CreateWallet(context.Background(), toWallet)
	assert.NoError(t, err)

	got, err := acc.Swap(context.Background(), types.SwapRequestOpts{
		CustomerID:       "cust_10",
		FromCurrencyCode: "ngn",
		ToCurrencyCode:   "usd",
		FromAmount:       decimal.NewFromFloat(1500),
		FromFee:          decimal.NewFromFloat(0),
		ToAmount:         decimal.NewFromFloat(1),
	})

	assert.NoError(t, err)
	assert.Equal(t, decimal.Zero.String(), got.FromWallet.AvailableBalance.String())
	assert.Equal(t, decimal.NewFromFloat(1).String(), got.ToWallet.AvailableBalance.String())
	assert.Equal(t, types.TransactionTypeSwap, got.FromTransaction.Type)
	assert.Equal(t, types.TransactionTypeSwap, got.ToTransaction.Type)
	assert.Nil(t, got.FromTransaction.Narration)
	assert.Nil(t, got.ToTransaction.Narration)

	// t.Errorf("%+v \n %+v \n %+v \n %+v", got.FromWallet, got.ToWallet, got.ToTransaction, got.FromTransaction)
}
