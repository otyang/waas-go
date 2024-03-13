package waas

import (
	"context"
	"errors"
	"time"

	gonanoid "github.com/matoous/go-nanoid"
)

// Defines a new error type for common errors encountered in the wallet module.
type WaasError struct{ error }

func NewWaasError(msg string) error { return &WaasError{errors.New(msg)} }

// IsWaasError checks if an error is of type WaasError.
func IsWaasError(err error) bool {
	_, ok := err.(*WaasError)
	return ok
}

func NewTransactionID() string {
	return time.Now().UTC().Format("20060102") + "_" + GenerateID(6)
}

func GenerateID(size int) string {
	// time.Now().UTC().Format("2006-01-02 15:04:05.999999")
	return gonanoid.MustGenerate("0123456789abcdefghijklmnopqrstuvwxyz", size)
}

type (
	TransactionType   string
	TransactionStatus string
)

const (
	TransactionTypeSwap       TransactionType = "SWAP"
	TransactionTypeTransfer   TransactionType = "TRANSFER"
	TransactionTypeDeposit    TransactionType = "DEPOSIT"
	TransactionTypeWithdrawal TransactionType = "WITHDRAWAL"
)

const (
	TransactionStatusNew     TransactionStatus = "NEW"
	TransactionStatusPending TransactionStatus = "PENDING"
	TransactionStatusFailed  TransactionStatus = "FAILED"
	TransactionStatusSuccess TransactionStatus = "SUCCESS"
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
