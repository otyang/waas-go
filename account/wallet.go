package account

import (
	"context"

	"github.com/otyang/waas-go"
)

func (a *Account) CreateWallet(ctx context.Context, wallet *waas.Wallet) (*waas.Wallet, error) {
	return nil, nil
}

func (a *Account) GetWalletByID(ctx context.Context, walletID string) (*waas.Wallet, error) {
	return nil, nil
}

func (a *Account) GetWalletByUserIDAndCurrencyCode(ctx context.Context, userID, currencyCode string) (*waas.Wallet, error) {
	return nil, nil
}

func (a *Account) UpdateWallet(ctx context.Context, wallet *waas.Wallet) (*waas.Wallet, error) {
	return nil, nil
}

func (a *Account) ListWallet(ctx context.Context, params waas.ListWalletsFilterParams) ([]waas.Wallet, error) {
	return nil, nil
}
