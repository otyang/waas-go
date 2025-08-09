package types

import (
	"time"

	"github.com/google/uuid"
)

// FreezeRequest contains details for freezing a wallet
type FreezeRequest struct {
	Reason      string `json:"reason"`      // Reason for freezing
	InitiatedBy string `json:"initiatedBy"` // Who requested the freeze
	ReferenceID string `json:"referenceId"` // External reference ID
}

// FreezeResult contains information about the freeze operation
type FreezeResult struct {
	WalletID  string    `json:"walletId"`  // Frozen wallet ID
	FreezeID  string    `json:"freezeId"`  // Unique freeze operation ID
	FrozenAt  time.Time `json:"frozenAt"`  // When freeze was applied
	Reason    string    `json:"reason"`    // Freeze reason
	ExpiresAt time.Time `json:"expiresAt"` // When freeze will auto-expire
}

// Freeze locks a wallet from all debit transactions while allowing credits
func (w *Wallet) Freeze(req FreezeRequest) (*FreezeResult, error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.IsClosed {
		return nil, ErrWalletClosed
	}
	if w.Frozen {
		return nil, ErrWalletAlreadyFrozen
	}

	w.Frozen = true
	w.FreezeReason = req.Reason
	w.FreezeInitiatedBy = req.InitiatedBy
	freezeTime := time.Now()
	w.FrozenAt = freezeTime
	w.UpdatedAt = freezeTime

	return &FreezeResult{
		WalletID: w.ID,
		FreezeID: uuid.New().String(),
		FrozenAt: freezeTime,
		Reason:   req.Reason,
	}, nil
}

// UnfreezeRequest contains details for unfreezing a wallet
type UnfreezeRequest struct {
	Reason      string `json:"reason"`      // Reason for unfreezing
	InitiatedBy string `json:"initiatedBy"` // Who requested the unfreeze
	ReferenceID string `json:"referenceId"` // External reference ID
}

// UnfreezeResult contains information about the unfreeze operation
type UnfreezeResult struct {
	WalletID       string        `json:"walletId"`       // Unfrozen wallet ID
	UnfreezeID     string        `json:"unfreezeId"`     // Unique unfreeze operation ID
	UnfrozenAt     time.Time     `json:"unfrozenAt"`     // When unfreeze was applied
	Reason         string        `json:"reason"`         // Unfreeze reason
	FreezeDuration time.Duration `json:"freezeDuration"` // How long wallet was frozen
}

// Unfreeze removes all transaction restrictions from a wallet
func (w *Wallet) Unfreeze(req UnfreezeRequest) (*UnfreezeResult, error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if !w.Frozen {
		return nil, ErrWalletNotFrozen
	}

	unfreezeTime := time.Now()
	freezeDuration := unfreezeTime.Sub(w.FrozenAt)

	// Clear all freeze-related fields
	w.Frozen = false
	w.FreezeReason = ""
	w.FreezeInitiatedBy = ""
	w.FrozenAt = time.Time{}
	w.UpdatedAt = unfreezeTime

	return &UnfreezeResult{
		WalletID:       w.ID,
		UnfreezeID:     uuid.New().String(),
		UnfrozenAt:     unfreezeTime,
		Reason:         req.Reason,
		FreezeDuration: freezeDuration,
	}, nil
}

// IsFrozen checks if wallet is currently frozen
func (w *Wallet) IsFrozen() bool {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	return w.Frozen
}

// GetFreezeInfo returns current freeze status
func (w *Wallet) GetFreezeInfo() *FreezeResult {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	if !w.Frozen {
		return nil
	}
	return &FreezeResult{
		WalletID: w.ID,
		FrozenAt: w.FrozenAt,
		Reason:   w.FreezeReason,
	}
}
