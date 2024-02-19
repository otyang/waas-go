package account

import (
	"context"

	"github.com/otyang/waas-go"
	"github.com/uptrace/bun"
)

type IAccountFeature interface {
	WalletCreate(ctx context.Context, wallet waas.Wallet) (*waas.Wallet, error)
	WalletGetByID(ctx context.Context, walletID string) (*waas.Wallet, error)
	WalletUpdate(ctx context.Context, wallet waas.Wallet) (*waas.Wallet, error)
	WalletList(ctx context.Context, params waas.ListWalletsFilterParams) ([]waas.Wallet, error)
	WalletGetByUserIDAndCurrencyCode(ctx context.Context, userID, currencyCode string) (*waas.Wallet, error)

	// Transaction Management
	TransactionCreate(ctx context.Context, transaction waas.Transaction) (*waas.Transaction, error)
	TransactionGetByID(ctx context.Context, transactionID string) (*waas.Transaction, error)
	TransactionUpdate(ctx context.Context, transaction waas.Transaction) (*waas.Transaction, error)
	TransactionList(ctx context.Context, params waas.ListTransactionsFilterParams) ([]waas.Transaction, error)

	Credit(ctx context.Context, params waas.CreditWalletParams) (*waas.CreditWalletResponse, error)
	Debit(ctx context.Context, params waas.DebitWalletParams) (*waas.DebitWalletResponse, error)
	Swap(ctx context.Context, params waas.SwapRequestParams) (*waas.SwapWalletResponse, error)
	Transfer(ctx context.Context, params waas.TransferRequestParams) (*waas.TransferResponse, error)
	Reverse(ctx context.Context, transactionID string) (*waas.ReverseResponse, error)
}

type Account struct {
	db *bun.IDB
}

func New(db *bun.DB) (Account, error)

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
