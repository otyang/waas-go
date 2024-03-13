package waas

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
)

// IWalletRepository defines repository functions for managing wallets and transactions.
type IAccountFeature interface {
	// Currency
	CreateCurrency(ctx context.Context, currency Currency) (*Currency, error)
	UpdateCurrency(ctx context.Context, currency Currency) (*Currency, error)
	ListCurrencies(ctx context.Context) ([]Currency, error)

	// wallets
	CreateWallet(ctx context.Context, wallet *Wallet) (*Wallet, error)
	GetWalletByID(ctx context.Context, walletID string) (*Wallet, error)
	GetWalletByUserIDAndCurrencyCode(ctx context.Context, userID, currencyCode string) (*Wallet, error)
	UpdateWallet(ctx context.Context, wallet *Wallet) (*Wallet, error)
	ListWallet(ctx context.Context, params ListWalletsFilterParams) ([]Wallet, error)

	// Transaction Management
	CreateTransaction(ctx context.Context, transaction *Transaction) (*Transaction, error)
	GetTransaction(ctx context.Context, transactionID string) (*Transaction, error)
	UpdateTransaction(ctx context.Context, transaction *Transaction) (*Transaction, error)
	ListTransaction(ctx context.Context, limit int, params ListTransactionsFilterParams) ([]Transaction, error)

	// actions
	Credit(ctx context.Context, params CreditWalletParams) (*CreditWalletResponse, error)
	Debit(ctx context.Context, params DebitWalletParams) (*DebitWalletResponse, error)
	Swap(ctx context.Context, params SwapRequestParams) (*SwapWalletResponse, error)
	Transfer(ctx context.Context, params TransferRequestParams) (*TransferResponse, error)
	Reverse(ctx context.Context, transactionID string) (*ReverseResponse, error)
}

type (
	// CreditWalletParams defines parameters for crediting a wallet.
	CreditWalletParams struct {
		WalletID  string
		Amount    decimal.Decimal
		Fee       decimal.Decimal
		Type      TransactionType
		Narration string
	}

	CreditWalletResponse struct {
		Wallet      *Wallet
		Transaction *Transaction
	}

	// DebitWalletParams defines parameters for debiting a wallet.
	DebitWalletParams struct {
		WalletID  string
		Amount    decimal.Decimal
		Fee       decimal.Decimal
		Type      TransactionType
		Status    TransactionStatus
		Narration string
	}

	DebitWalletResponse struct {
		Wallet      *Wallet
		Transaction *Transaction
	}

	// TransferParams defines parameters for transferring funds between wallets.
	TransferRequestParams struct {
		FromWalletID string          `json:"fromWid"`
		ToWalletID   string          `json:"toWid"`
		Amount       decimal.Decimal `json:"amount"`
		Fee          decimal.Decimal `json:"fee"`
		Narration    string          `json:"narration"`
	}

	TransferResponse struct {
		FromWallet      *Wallet
		ToWallet        *Wallet
		FromTransaction *Transaction
		ToTransaction   *Transaction
	}

	// ReverseParams defines parameters for reversing a transaction.
	ReverseRequestParams struct {
		TransactionID string `json:"transactionId"`
	}

	ReverseResponse struct {
		OldTx  *Transaction
		NewTx  *Transaction
		Wallet *Wallet
	}

	// SwapParams defines parameters for swapping currencies between wallets.
	SwapRequestParams struct {
		UserID           string
		FromCurrencyCode string
		ToCurrencyCode   string
		FromAmount       decimal.Decimal
		FromFee          decimal.Decimal
		ToAmount         decimal.Decimal
	}

	SwapWalletResponse struct {
		FromWallet      *Wallet
		ToWallet        *Wallet
		FromTransaction *Transaction
		ToTransaction   *Transaction
	}

	ListWalletsFilterParams struct {
		CustomerID   *string
		CurrencyCode *string
		IsFiat       *bool
		IsFrozen     *bool
	}

	ListTransactionsFilterParams struct {
		After      time.Time
		Before     time.Time
		CustomerID *string
		WalletID   *string
		Currency   *string
		IsDebit    *bool
		Type       *TransactionType
		Status     *TransactionStatus
		Reversed   *bool
	}
)
