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
