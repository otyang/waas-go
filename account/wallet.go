package account

import (
	"context"
	"time"

	"github.com/otyang/waas-go"
)

func (a *Account) CreateWallet(ctx context.Context, wallet *waas.Wallet) (*waas.Wallet, error) {
	wallet.CreatedAt = time.Now()
	wallet.UpdatedAt = time.Now()
	_, err := a.db.NewInsert().Model(wallet).Ignore().Exec(ctx)
	return wallet, err
}

func (a *Account) GetWalletByID(ctx context.Context, walletID string) (*waas.Wallet, error) {
	wallet := waas.Wallet{ID: walletID}
	err := a.db.NewSelect().Model(&wallet).WherePK().Limit(1).Scan(ctx)
	return &wallet, err
}

func (acc *Account) GetWalletByUserIDAndCurrencyCode(ctx context.Context, userID, currencyCode string) (*waas.Wallet, error) {
	return acc.GetWalletByID(ctx, waas.GenerateWalletID(currencyCode, userID))
}

func (a *Account) UpdateWallet(ctx context.Context, wallet *waas.Wallet) (*waas.Wallet, error) {
	oldVersionID := wallet.VersionId      // extract oldVersionID. for concurrency locks
	wallet.VersionId = waas.GenerateID(7) // newVId
	wallet.UpdatedAt = time.Now()

	_, err := a.db.NewUpdate().Model(wallet).WherePK().Where("version_id = ?", oldVersionID).Exec(ctx)

	return wallet, err
}

func (a *Account) ListWallet(ctx context.Context, params waas.ListWalletsFilterParams) ([]waas.Wallet, error) {
	var wallets []waas.Wallet

	q := a.db.NewSelect().Model(&wallets)

	if params.CustomerID != nil {
		q.Where("customer_id = ?", params.CustomerID)
	}

	if params.CurrencyCode != nil {
		q.Where("currency_code = ?", params.CurrencyCode)
	}

	if params.IsFiat != nil {
		q.Where("is_fiat = ?", params.IsFiat)
	}

	if params.IsFrozen != nil {
		q.Where("is_frozen = ?", params.IsFrozen)
	}

	err := q.Scan(ctx)
	return wallets, err
}
