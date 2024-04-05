package account

import (
	"context"
	"time"

	"github.com/otyang/waas-go/types"
	"github.com/uptrace/bun"
)

func (a *Account) CreateWallet(ctx context.Context, wallet *types.Wallet) (*types.Wallet, error) {
	wallet.CreatedAt = time.Now()
	wallet.UpdatedAt = time.Now()
	_, err := a.db.NewInsert().Model(wallet).Ignore().Exec(ctx)
	return wallet, err
}

func (a *Account) GetWalletByID(ctx context.Context, walletID string) (*types.Wallet, error) {
	wallet := types.Wallet{ID: walletID}
	err := a.db.NewSelect().Model(&wallet).WherePK().Limit(1).Scan(ctx)
	return &wallet, err
}

func (acc *Account) GetWalletByUserIDAndCurrencyCode(ctx context.Context, userID, currencyCode string) (*types.Wallet, error) {
	return acc.GetWalletByID(ctx, types.GenerateWalletID(currencyCode, userID))
}

func (a *Account) UpdateWallet(ctx context.Context, wallet *types.Wallet) (*types.Wallet, error) {
	oldVersionID := wallet.VersionId       // extract oldVersionID. for concurrency locks
	wallet.VersionId = types.GenerateID(7) // newVId
	wallet.UpdatedAt = time.Now()

	_, err := a.db.NewUpdate().Model(wallet).WherePK().Where("version_id = ?", oldVersionID).Exec(ctx)
	return wallet, err
}

func (a *Account) ListWallet(ctx context.Context, params types.ListWalletsFilterParams) ([]types.Wallet, error) {
	var wallets []types.Wallet

	q := a.db.NewSelect().Model(&wallets)

	if params.CustomerID != nil {
		q.Where("customer_id = ?", params.CustomerID)
	}

	if len(params.CurrencyCodes) > 0 {
		q.Where("lower(currency_code) IN (?)", bun.In(types.ToLowercaseSlice(params.CurrencyCodes)))
	}

	if params.Status != nil {
		q.Where("lower(status) = lower(?)", params.Status)
	}

	err := q.Order("currency_code ASC").Scan(ctx)
	return wallets, err
}
