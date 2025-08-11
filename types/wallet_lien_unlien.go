package types

import (
	"errors"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

// Error definitions
var (
	ErrInvalidLienID = errors.New("invalid lien ID")
)

// Lien contains details for placing/releasing liens
type LienOrUnlienRequest struct {
	ID          string          `json:"id"`          // Unique ID
	Amount      decimal.Decimal `json:"amount"`      // Positive amount to lien/unlien
	Description string          `json:"description"` // Context for the operation
}

// LienRecord contains the complete record of a lien operation
type LienRecord struct {
	ID          string          `json:"referenceId"` // Unique reference ID
	WalletID    string          `json:"walletId"`    // Affected wallet ID
	Amount      decimal.Decimal `json:"amount"`      // Amount liened/released
	Description string          `json:"description"` // Operation context
	CreatedAt   time.Time       `json:"createdAt"`   // When lien was placed
	ReleasedAt  time.Time       `json:"releasedAt"`  // When lien was released (if applicable)
}

// AddLien places a lien on the specified amount from available balance
func (w *Wallet) AddLien(lien LienOrUnlienRequest) (*LienRecord, error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	// Validate wallet state
	if w.IsClosed {
		return nil, ErrWalletClosed
	}
	if w.Frozen {
		return nil, ErrWalletFrozen
	}

	// Validate lien amount
	if lien.Amount.LessThanOrEqual(decimal.Zero) {
		return nil, ErrInvalidAmount
	}

	// Check sufficient available balance
	if w.AvailableBalance.LessThan(lien.Amount) {
		return nil, ErrInsufficientFunds
	}

	// Update balances
	w.AvailableBalance = w.AvailableBalance.Sub(lien.Amount)
	w.LienBalance = w.LienBalance.Add(lien.Amount)
	now := time.Now()
	w.UpdatedAt = now

	// Create and return lien record
	return &LienRecord{
		ID:          GenerateID("lien_", 15),
		WalletID:    w.ID,
		Amount:      lien.Amount,
		Description: lien.Description,
		CreatedAt:   now,
	}, nil
}

// ReleaseLien releases a lien and returns funds to available balance
func (w *Wallet) ReleaseLien(lien LienOrUnlienRequest) (*LienRecord, error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	// Validate lien ID
	if strings.TrimSpace(lien.ID) == "" {
		return nil, ErrInvalidLienID
	}

	// Validate wallet state
	if w.IsClosed {
		return nil, ErrWalletClosed
	}

	// Validate lien amount
	if lien.Amount.LessThanOrEqual(decimal.Zero) {
		return nil, ErrInvalidAmount
	}

	// Check sufficient lien balance
	if w.LienBalance.LessThan(lien.Amount) {
		return nil, ErrInsufficientLien
	}

	// Update balances
	w.LienBalance = w.LienBalance.Sub(lien.Amount)
	w.AvailableBalance = w.AvailableBalance.Add(lien.Amount)
	now := time.Now()
	w.UpdatedAt = now

	// Return lien record with release timestamp
	return &LienRecord{
		ID:          lien.ID,
		WalletID:    w.ID,
		Amount:      lien.Amount,
		Description: lien.Description,
		ReleasedAt:  now,
	}, nil
}
