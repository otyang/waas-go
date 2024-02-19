package waas

import (
	"context"

	"github.com/shopspring/decimal"
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

type (
	// CreditWalletParams defines parameters for crediting a wallet.
	CreditWalletParams struct {
		WalletID        string
		Amount          decimal.Decimal
		Fee             decimal.Decimal
		Type            TransactionType
		SourceNarration string
	}

	CreditWalletResponse struct {
		AmountTransfered decimal.Decimal
		Fee              decimal.Decimal
		Wallet           Wallet
		Transaction      Transaction
	}

	// DebitWalletParams defines parameters for debiting a wallet.
	DebitWalletParams struct {
		WalletID        string
		Amount          decimal.Decimal
		Fee             decimal.Decimal
		Type            TransactionType
		SourceNarration string
	}

	DebitWalletResponse struct {
		AmountTransfered decimal.Decimal
		Fee              decimal.Decimal
		Wallet           Wallet
		Transaction      Transaction
	}

	// TransferParams defines parameters for transferring funds between wallets.
	TransferRequestParams struct {
		FromWalletID    string          `json:"fromWid"`
		ToWalletID      string          `json:"toWid"`
		Amount          decimal.Decimal `json:"amount"`
		Fee             decimal.Decimal `json:"fee"`
		SourceNarration string
	}

	TransferResponse struct {
		AmountTransfered decimal.Decimal
		Fee              decimal.Decimal
		FromWallet       Wallet
		ToWallet         Wallet
		FromTransaction  Transaction
		ToTransaction    Transaction
	}

	// ReverseParams defines parameters for reversing a transaction.
	ReverseRequestParams struct {
		TransactionID string `json:"transactionId"`
	}

	ReverseResponse struct {
		OldUpdatedTx  *Transaction
		NewTx         *Transaction
		UpdatedWallet *Wallet
	}

	// SwapParams defines parameters for swapping currencies between wallets.
	SwapRequestParams struct {
		UserID           string
		FromCurrencyCode string
		ToCurrencyCode   string
		Amount           decimal.Decimal
		Fee              decimal.Decimal
	}

	SwapWalletResponse struct {
		FromWallet      Wallet
		ToWallet        Wallet
		FromTransaction Transaction
		ToTransaction   Transaction
	}

	ListWalletsFilterParams struct {
		CustomerID   *string
		CurrencyCode *string
		IsFiat       *bool
		IsFrozen     bool
	}

	ListTransactionsFilterParams struct {
		CustomerID string
		Currency   string
		IsDebit    *bool
		Type       *TransactionType
		Status     *TransactionStatus
		Narration  string
		Reversed   *bool
	}
)
