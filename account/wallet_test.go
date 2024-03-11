package account

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAccount_CreateWallet(t *testing.T) {
	acc := &Account{
		db: TestDB,
	}

	w := createTestRandomWallet("cust_123", "ngn")

	got, err := acc.CreateWallet(context.Background(), w)
	assert.NoError(t, err)
	assert.Equal(t, w, got)

	// Attempt creating same wallet again to ensure its ignored
	got, err = acc.CreateWallet(context.Background(), got)
	assert.NoError(t, err)
	assert.Equal(t, w.VersionId, got.VersionId) // if new is created/updated VersionID will be different

	// Get wallet
	got, err = acc.GetWalletByID(context.Background(), w.ID)
	assert.NoError(t, err)
	assert.Equal(t, got.ID, got.ID)
	assert.Equal(t, got.VersionId, got.VersionId)

	// Get wallet by Id & CurrencyCode
	got, err = acc.GetWalletByUserIDAndCurrencyCode(context.Background(), "cust_123", "ngn")
	assert.NoError(t, err)
	assert.Equal(t, got.VersionId, got.VersionId)

	// Update wallet
	got.IsFiat = true
	v := got.VersionId
	actual, err := acc.UpdateWallet(context.Background(), got)
	assert.NoError(t, err)
	assert.NotEqual(t, v, actual.VersionId)

	// waas.ListWalletsFilterParams
	// got, err := a.ListWallet(tt.args.ctx, tt.args.params)
}
