package account

import (
	"context"
	"database/sql"
	"time"

	"github.com/otyang/waas-go/types"
	"github.com/uptrace/bun"
)

// =get walletandcurrency

func (a *Client) CreateWallet(ctx context.Context, wallet *types.Wallet) (*types.Wallet, error) {
	foundWallet, err := a.GetWalletByCurrencyCode(ctx, wallet.CurrencyCode, wallet.CustomerID)
	if err == nil {
		return foundWallet, nil
	}

	if err == sql.ErrNoRows {
		_, err = a.db.NewInsert().Model(wallet).Ignore().Exec(ctx)
		return wallet, err
	}
	return nil, err
}

func (a *Client) CreateWalletSimplified(ctx context.Context, customerID, currencyCode string) (*types.Wallet, error) {
	return a.CreateWallet(ctx, types.NewWallet(customerID, currencyCode))
}

func (a *Client) UpdateWallet(ctx context.Context, wallet *types.Wallet) (*types.Wallet, error) {
	oldVID := wallet.VersionId // for concurrency locks
	wallet.UpdatedAt = time.Now()
	wallet.VersionId = types.GenerateID("v_", 7) // newVID

	_, err := a.db.
		NewUpdate().
		Model(wallet).WherePK().
		Where("version_id = ?", oldVID).Exec(ctx)
	return wallet, err
}

func (a *Client) GetWalletByID(ctx context.Context, walletID string) (*types.Wallet, error) {
	wallet := types.Wallet{ID: walletID}
	err := a.db.
		NewSelect().
		Model(&wallet).
		WherePK().Limit(1).Scan(ctx)
	return &wallet, err
}

func (a *Client) GetWalletByCurrencyCode(ctx context.Context, currencyCode, customerID string) (*types.Wallet, error) {
	var wallet types.Wallet

	err := a.db.NewSelect().
		Model(&wallet).
		Where("customer_id = ?", customerID).
		Where("lower(currency_code) = lower(?)", currencyCode).Limit(1).Scan(ctx)

	return &wallet, err
}

func (a *Client) ListWallets(ctx context.Context, opts types.ListWalletsFilterOpts) ([]types.Wallet, error) {
	var wallets []types.Wallet

	q := a.db.NewSelect().Model(&wallets)

	if opts.CustomerID != "" {
		q.Where("customer_id = ?", opts.CustomerID)
	}

	if len(opts.CurrencyCodes) > 0 {
		q.Where(
			"lower(currency_code) IN (?)", bun.In(types.ToLowercaseSlice(opts.CurrencyCodes)),
		)
	}

	if opts.IsFrozen != nil {
		q.Where("lower(is_frozen) = ?", opts.IsFrozen)
	}

	if opts.IsClosed != nil {
		q.Where("lower(is_closed) = ?", opts.IsClosed)
	}

	err := q.Order("currency_code ASC").Scan(ctx)
	return wallets, err
}
