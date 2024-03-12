package account

import (
	"context"
	"testing"

	"github.com/otyang/waas-go"
	"github.com/stretchr/testify/assert"
)

func TestAccount_Wallets_and_all_its_assosiated_functions(t *testing.T) {
	acc := &Account{
		db: TestDB,
	}

	w := createTestRandomWallet("cust_123", "ngn")

	got, err := acc.CreateWallet(context.Background(), w)
	assert.NoError(t, err)
	assert.Equal(t, w, got)

	t.Run("Re-Create same existing wallet. It should ignore", func(t *testing.T) {
		got, err = acc.CreateWallet(context.Background(), got)
		assert.NoError(t, err)
		assert.Equal(t, w.VersionId, got.VersionId) // if new is created/updated VersionID will be different
	})

	t.Run("Get wallet", func(t *testing.T) {
		got, err = acc.GetWalletByID(context.Background(), w.ID)
		assert.NoError(t, err)
		assert.Equal(t, got.ID, got.ID)
		assert.Equal(t, got.VersionId, got.VersionId)
	})

	t.Run("Get wallet by Id & CurrencyCode", func(t *testing.T) {
		got, err = acc.GetWalletByUserIDAndCurrencyCode(context.Background(), "cust_123", "ngn")
		assert.NoError(t, err)
		assert.Equal(t, got.VersionId, got.VersionId)
	})

	t.Run("Update wallet", func(t *testing.T) {
		got.IsFiat = true
		v := got.VersionId
		actual, err := acc.UpdateWallet(context.Background(), got)
		assert.NoError(t, err)
		assert.NotEqual(t, v, actual.VersionId)
	})

	// list wallets test. lets add more
	w1 := createTestRandomWallet("cust_123", "usd")
	_, err = acc.CreateWallet(context.Background(), w1)
	assert.NoError(t, err)

	t.Run("list without filters", func(t *testing.T) {
		gotList, err := acc.ListWallet(context.Background(), waas.ListWalletsFilterParams{})
		assert.NoError(t, err)
		assert.Equal(t, 2, len(gotList))
	})

	t.Run("list with filters", func(t *testing.T) {
		gotList, err := acc.ListWallet(context.Background(), waas.ListWalletsFilterParams{
			CustomerID:   toPointer("customerID"),
			CurrencyCode: toPointer("currencyCode"),
			IsFiat:       toPointer(true),
			IsFrozen:     toPointer(false),
		})
		assert.NoError(t, err)
		assert.Equal(t, 0, len(gotList))
	})
}
