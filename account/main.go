package account

import (
	"context"

	"github.com/otyang/waas-go"
	"github.com/otyang/waas-go/currency"
	"github.com/uptrace/bun"
)

type IAccountFeature interface {
	// Currency
	CreateCurrency(ctx context.Context, currency currency.Currency) (*currency.Currency, error)
	UpdateCurrency(ctx context.Context, currency currency.Currency) (*currency.Currency, error)
	ListCurrencies(ctx context.Context) ([]currency.Currency, error)

	// wallets
	CreateWallet(ctx context.Context, wallet *waas.Wallet) (*waas.Wallet, error)
	GetWalletByID(ctx context.Context, walletID string) (*waas.Wallet, error)
	GetWalletByUserIDAndCurrencyCode(ctx context.Context, userID, currencyCode string) (*waas.Wallet, error)
	UpdateWallet(ctx context.Context, wallet *waas.Wallet) (*waas.Wallet, error)
	ListWallet(ctx context.Context, params waas.ListWalletsFilterParams) ([]waas.Wallet, error)

	// Transaction Management
	CreateTransaction(ctx context.Context, transaction *waas.Transaction) (*waas.Transaction, error)
	GetTransaction(ctx context.Context, transactionID string) (*waas.Transaction, error)
	UpdateTransaction(ctx context.Context, transaction *waas.Transaction) (*waas.Transaction, error)
	ListTransaction(ctx context.Context, limit int, params waas.ListTransactionsFilterParams) ([]waas.Transaction, error)

	// actions
	Credit(ctx context.Context, params waas.CreditWalletParams) (*waas.CreditWalletResponse, error)
	Debit(ctx context.Context, params waas.DebitWalletParams) (*waas.DebitWalletResponse, error)
	Swap(ctx context.Context, params waas.SwapRequestParams) (*waas.SwapWalletResponse, error)
	Transfer(ctx context.Context, params waas.TransferRequestParams) (*waas.TransferResponse, error)
	Reverse(ctx context.Context, transactionID string) (*waas.ReverseResponse, error)
}

var _ IAccountFeature = (*Account)(nil)

type Account struct {
	db bun.IDB
}

func New(db *bun.DB) *Account {
	return &Account{db: db}
}

func (a *Account) NewWithTx(tx bun.Tx) *Account {
	return &Account{
		db: tx,
	}
}

func (a *Account) WithTxBulkUpdateWalletAndTransaction(ctx context.Context, wallets []*waas.Wallet, transactions []*waas.Transaction) error {
	return a.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		for _, wallet := range wallets {
			if wallet == nil {
				continue
			}
			_, err := a.NewWithTx(tx).UpdateWallet(ctx, wallet)
			if err != nil {
				return err
			}
		}

		for _, transaction := range transactions {
			if transaction == nil {
				continue
			}
			_, err := a.NewWithTx(tx).CreateTransaction(ctx, transaction)
			if err != nil {
				return err
			}
		}

		return nil
	})
}
