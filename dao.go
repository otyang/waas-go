package waas

import (
	"context"
)

// IWalletRepository defines repository functions for managing wallets and transactions.
type IAccountFeature interface {
	WalletCreate(ctx context.Context, wallet Wallet) (*Wallet, error)
	WalletGetByID(ctx context.Context, walletID string) (*Wallet, error)
	WalletUpdate(ctx context.Context, wallet Wallet) (*Wallet, error)
	WalletList(ctx context.Context, params ListWalletsFilterParams) ([]Wallet, error)
	WalletGetByUserIDAndCurrencyCode(ctx context.Context, userID, currencyCode string) (*Wallet, error)

	// Transaction Management
	TransactionCreate(ctx context.Context, transaction Transaction) (*Transaction, error)
	TransactionGetByID(ctx context.Context, transactionID string) (*Transaction, error)
	TransactionUpdate(ctx context.Context, transaction Transaction) (*Transaction, error)
	TransactionList(ctx context.Context, params ListTransactionsFilterParams) ([]Transaction, error)

	Credit(ctx context.Context, params CreditWalletParams) (*CreditWalletResponse, error)
	Debit(ctx context.Context, params DebitWalletParams) (*DebitWalletResponse, error)
	Swap(ctx context.Context, params SwapRequestParams) (*SwapWalletResponse, error)
	Transfer(ctx context.Context, params TransferRequestParams) (*TransferResponse, error)
	Reverse(ctx context.Context, transactionID string) (*ReverseResponse, error)
}
