package account

import (
	"context"
	"testing"

	"github.com/otyang/waas-go"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestAccount_Credit_and_Debit(t *testing.T) {
	t.Parallel()

	acc := &Account{
		db: TestDB,
	}

	w := createTestRandomWallet("cust_123", "ngn")

	_, err := acc.CreateWallet(context.Background(), w)
	assert.NoError(t, err)

	t.Run("crediting", func(t *testing.T) {
		got, err := acc.Credit(context.Background(), waas.CreditWalletParams{
			WalletID:  w.ID,
			Amount:    decimal.NewFromFloat(50),
			Fee:       decimal.Zero,
			Type:      waas.TransactionTypeDeposit,
			Narration: "deposit of funds",
		})

		assert.NoError(t, err)
		assert.Equal(t, got.Wallet.AvailableBalance.String(), decimal.NewFromFloat(50).String())
		assert.Equal(t, waas.TransactionTypeDeposit, got.Transaction.Type)
		assert.Equal(t, "deposit of funds", *got.Transaction.Narration)
	})

	t.Run("debiting", func(t *testing.T) {
		got, err := acc.Debit(context.Background(), waas.DebitWalletParams{
			WalletID:  w.ID,
			Amount:    decimal.NewFromFloat(20),
			Fee:       decimal.Zero,
			Type:      waas.TransactionTypeWithdrawal,
			Narration: "withdrawal of funds",
		})

		assert.NoError(t, err)
		assert.Equal(t, got.Wallet.AvailableBalance.String(), decimal.NewFromFloat(30).String())
		assert.Equal(t, waas.TransactionTypeWithdrawal, got.Transaction.Type)
		assert.Equal(t, "withdrawal of funds", *got.Transaction.Narration)
	})
}

func TestAccount_Transfer(t *testing.T) {
	t.Parallel()

	acc := &Account{
		db: TestDB,
	}

	sourceWallet := createTestRandomWallet("cust_1", "ngn")
	sourceWallet.AvailableBalance = decimal.NewFromFloat(30)
	destinationWallet := createTestRandomWallet("cust_2", "ngn")

	_, err := acc.CreateWallet(context.Background(), sourceWallet)
	assert.NoError(t, err)

	_, err = acc.CreateWallet(context.Background(), destinationWallet)
	assert.NoError(t, err)

	got, err := acc.Transfer(context.Background(), waas.TransferRequestParams{
		FromWalletID: sourceWallet.ID,
		ToWalletID:   destinationWallet.ID,
		Amount:       decimal.NewFromFloat(30),
		Fee:          decimal.Zero,
		Narration:    "transfer of funds",
	})

	assert.NoError(t, err)
	assert.Equal(t, decimal.Zero.String(), got.FromWallet.AvailableBalance.String())
	assert.Equal(t, decimal.NewFromFloat(30).String(), got.ToWallet.AvailableBalance.String())
	assert.Equal(t, waas.TransactionTypeTransfer, got.FromTransaction.Type)
	assert.Equal(t, waas.TransactionTypeTransfer, got.ToTransaction.Type)
	assert.Equal(t, "transfer of funds", *got.FromTransaction.Narration)
	assert.Equal(t, "transfer of funds", *got.ToTransaction.Narration)
}

func TestAccount_Swap(t *testing.T) {
	t.Parallel()

	acc := &Account{
		db: TestDB,
	}

	fromWallet := createTestRandomWallet("cust_10", "ngn")
	fromWallet.AvailableBalance = decimal.NewFromFloat(1500)
	toWallet := createTestRandomWallet("cust_10", "usd")

	_, err := acc.CreateWallet(context.Background(), fromWallet)
	assert.NoError(t, err)

	_, err = acc.CreateWallet(context.Background(), toWallet)
	assert.NoError(t, err)

	got, err := acc.Swap(context.Background(), waas.SwapRequestParams{
		UserID:           "cust_10",
		FromCurrencyCode: "ngn",
		ToCurrencyCode:   "usd",
		FromAmount:       decimal.NewFromFloat(1500),
		FromFee:          decimal.NewFromFloat(0),
		ToAmount:         decimal.NewFromFloat(1),
	})

	assert.NoError(t, err)
	assert.Equal(t, decimal.Zero.String(), got.FromWallet.AvailableBalance.String())
	assert.Equal(t, decimal.NewFromFloat(1).String(), got.ToWallet.AvailableBalance.String())
	assert.Equal(t, waas.TransactionTypeSwap, got.FromTransaction.Type)
	assert.Equal(t, waas.TransactionTypeSwap, got.ToTransaction.Type)
	assert.Nil(t, got.FromTransaction.Narration)
	assert.Nil(t, got.ToTransaction.Narration)

	// t.Errorf("%+v \n %+v \n %+v \n %+v", got.FromWallet, got.ToWallet, got.ToTransaction, got.FromTransaction)
}
