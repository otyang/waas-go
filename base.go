package waas

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
	GetWalletByUserIDAndCurrencyCode(ctx context.Context, userID, currencyCode string) (*Wallet, error)
	UpdateWallet(ctx context.Context, wallet *Wallet) (*Wallet, error)
	ListWallet(ctx context.Context, params ListWalletsFilterParams) ([]Wallet, error)

	// Transaction Management
	CreateTransaction(ctx context.Context, transaction *Transaction) (*Transaction, error)
	GetTransaction(ctx context.Context, transactionID string) (*Transaction, error)
	UpdateTransaction(ctx context.Context, transaction *Transaction) (*Transaction, error)
	ListTransaction(ctx context.Context, params ListTransactionsFilterParams) ([]Transaction, string, error)

	// actions
	Credit(ctx context.Context, params CreditWalletParams) (*CreditWalletResponse, error)
	Debit(ctx context.Context, params DebitWalletParams) (*DebitWalletResponse, error)
	Swap(ctx context.Context, params SwapRequestParams) (*SwapWalletResponse, error)
	Transfer(ctx context.Context, params TransferRequestParams) (*TransferResponse, error)
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
	return time.Now().UTC().Format("20060102") + "_" + GenerateID(8)
}

func GenerateID(size int) string {
	// time.Now().UTC().Format("2006-01-02 15:04:05.999999")
	return gonanoid.MustGenerate("0123456789abcdefghijklmnopqrstuvwxyz", size)
}

// helps in making the in sql case insensitive
func ToLowercaseSlice(strs []string) []string {
	lowercaseStrs := make([]string, len(strs))

	for i, str := range strs {
		lowercaseStrs[i] = strings.TrimSpace(strings.ToLower(str)) // Convert each string
	}

	return lowercaseStrs
}
