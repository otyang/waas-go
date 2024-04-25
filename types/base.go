package types

import (
	"context"
	"strings"
	"time"

	gonanoid "github.com/matoous/go-nanoid"
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
	GetWalletByCurrencyCode(ctx context.Context, userID, currencyCode string) (*Wallet, error)
	UpdateWallet(ctx context.Context, wallet *Wallet) (*Wallet, error)
	ListWallets(ctx context.Context, opts ListWalletsFilterOpts) ([]Wallet, error)

	// Transaction Management
	CreateTransaction(ctx context.Context, transaction *Transaction) (*Transaction, error)
	GetTransaction(ctx context.Context, transactionID string) (*Transaction, error)
	UpdateTransaction(ctx context.Context, transaction *Transaction) (*Transaction, error)
	ListTransactions(ctx context.Context, opts ListTransactionsFilterOpts) ([]Transaction, string, error)

	// actions
	Credit(ctx context.Context, opts CreditWalletOpts) (*CreditWalletResponse, error)
	Debit(ctx context.Context, opts DebitWalletOpts) (*DebitWalletResponse, error)
	Swap(ctx context.Context, opts SwapRequestOpts) (*SwapWalletResponse, error)
	Transfer(ctx context.Context, opts TransferRequestOpts) (*TransferResponse, error)
	Reverse(ctx context.Context, transactionID string) (*ReverseResponse, error)
}

// WalletError provides a common base for wallet-related errors.
type WaasError struct{ msg string }

func (e *WaasError) Error() string { return e.msg }

// NewWaasError creates a new WalletError.
func NewWaasError(msg string) *WaasError {
	return &WaasError{msg}
}

// IsWaasError checks if an error is of type WaasError.
func IsWaasError(err error) bool {
	_, ok := err.(*WaasError)
	return ok
}

func NewTransactionID() string {
	return GenerateID("txn_"+time.Now().UTC().Format("20060102")+"_", 8)
}

func GenerateID(prefix string, size int) string {
	return prefix + gonanoid.MustGenerate("0123456789abcdefghijklmnopqrstuvwxyz", size)
}

// makes a slice of strings insensitive
func ToLowercaseSlice(strs []string) []string {
	lowercaseStrs := make([]string, len(strs))
	for i, str := range strs {
		lowercaseStrs[i] = strings.TrimSpace(strings.ToLower(str))
	}
	return lowercaseStrs
}
