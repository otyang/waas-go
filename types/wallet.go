package types

import (
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/shopspring/decimal"
)

// Wallet holds funds for a customer with thread-safe operations
type Wallet struct {
	ID                string          `json:"id"`                     // Unique ID
	CustomerID        string          `json:"customerId"`             // Owner ID
	AvailableBalance  decimal.Decimal `json:"availableBalance"`       // Spendable amount
	LienBalance       decimal.Decimal `json:"lienBalance"`            // Reserved amount
	CurrencyCode      string          `json:"currencyCode"`           // Currency type (USD, EUR etc.)
	IsClosed          bool            `json:"isClosed"`               // Closed flag
	Frozen            bool            `json:"frozen"`                 // Frozen flag
	FreezeReason      string          `json:"freezeReason,omitempty"` // Freeze Reason
	FreezeInitiatedBy string          `json:"freezeInitiatedBy"`
	FrozenAt          time.Time       `json:"frozenAt"`
	CreatedAt         time.Time       `json:"createdAt"` // Creation time
	UpdatedAt         time.Time       `json:"updatedAt"` // Last update time
	VersionId         string          `json:"-"`         // For concurrency control
	mutex             sync.RWMutex    `json:"-"`         // Thread safety
}

// NewWallet creates and initializes a new Wallet instance
func NewWallet(customerID, currencyCode string) (*Wallet, error) {
	// Validate inputs
	if customerID == "" {
		return nil, errors.New("customer ID cannot be empty")
	}
	if currencyCode == "" {
		return nil, errors.New("currency code cannot be empty")
	}

	now := time.Now().UTC()

	wallet := &Wallet{
		ID:               GenerateID("wt_", 10), // Generate a new UUID
		CustomerID:       customerID,
		AvailableBalance: decimal.NewFromInt(0),
		LienBalance:      decimal.NewFromInt(0),
		CurrencyCode:     strings.ToUpper(currencyCode),
		IsClosed:         false,
		Frozen:           false,
		FreezeReason:     "",
		FrozenAt:         time.Time{}, // Zero time
		//	FreezeInitiatedBy: stringPtr(""),
		CreatedAt: now,
		UpdatedAt: now,
		VersionId: GenerateID("wtvid_", 15), // Initial version
		// mutex is zero-valued which is fine for sync.RWMutex
	}

	return wallet, nil
}

// CanBeDebited checks if wallet is in a state that allows debits
func (w *Wallet) CanBeDebited() error {
	if w.IsClosed {
		return ErrWalletClosed
	}
	if w.Frozen {
		return ErrWalletFrozen
	}
	return nil
}

// CanBeCredited checks if wallet is in a state that allows credits
func (w *Wallet) CanBeCredited() error {
	if w.IsClosed {
		return ErrWalletClosed
	}
	return nil
}

// TotalBalance returns sum of available and lien balances
func (w *Wallet) TotalBalance() decimal.Decimal {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	return w.AvailableBalance.Add(w.LienBalance)
}
