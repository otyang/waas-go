package account

import (
	"context"
	"time"

	"github.com/otyang/waas-go/types"
	"github.com/uptrace/bun"
)

// create wallet, find wallet doesnt exist. then create or
func (a *Client) CreateWallet(ctx context.Context, wallet *types.Wallet) (*types.Wallet, error) {
	wallet.CreatedAt = time.Now()
	wallet.UpdatedAt = time.Now()
	_, err := a.db.NewInsert().Model(wallet).Ignore().Exec(ctx)
	return wallet, err
}

func (a *Client) GetWalletByID(ctx context.Context, walletID string) (*types.Wallet, error) {
	wallet := types.Wallet{ID: walletID}
	err := a.db.NewSelect().Model(&wallet).WherePK().Limit(1).Scan(ctx)
	return &wallet, err
}

func (acc *Client) GetWalletByCurrencyCode(ctx context.Context, userID, currencyCode string) (*types.Wallet, error) {
	return acc.GetWalletByID(ctx, types.GenerateWalletID(currencyCode, userID))
}

func (a *Client) UpdateWallet(ctx context.Context, wallet *types.Wallet) (*types.Wallet, error) {
	oldVersionID := wallet.VersionId       // extract oldVersionID. for concurrency locks
	wallet.VersionId = types.GenerateID(7) // newVId
	wallet.UpdatedAt = time.Now()

	_, err := a.db.NewUpdate().Model(wallet).WherePK().Where("version_id = ?", oldVersionID).Exec(ctx)
	return wallet, err
}

func (a *Client) ListWallets(ctx context.Context, opts types.ListWalletsFilterOpts) ([]types.Wallet, error) {
	var wallets []types.Wallet

	q := a.db.NewSelect().Model(&wallets)

	if opts.CustomerID != "" {
		q.Where("customer_id = ?", opts.CustomerID)
	}

	if len(opts.CurrencyCodes) > 0 {
		q.Where("lower(currency_code) IN (?)", bun.In(types.ToLowercaseSlice(opts.CurrencyCodes)))
	}

	if string(opts.Status) != "" {
		q.Where("lower(status) = lower(?)", opts.Status)
	}

	err := q.Order("currency_code ASC").Scan(ctx)
	return wallets, err
}
