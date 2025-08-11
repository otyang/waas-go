package types

import (
	"time"
)

// FreezeRequest contains details for freezing a wallet
type FreezeRequest struct {
	Reason      string    `json:"reason"`      // Reason for freezing
	InitiatedBy string    `json:"initiatedBy"` // Who requested the freeze
	FrozenAt    time.Time `json:"frozenAt"`    // When the freeze was applied
}

func (w *Wallet) Freeze(req FreezeRequest) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.IsClosed {
		return ErrWalletClosed
	}
	if w.Frozen {
		return ErrWalletAlreadyFrozen
	}

	w.Frozen = true
	w.FreezeReason = req.Reason
	w.FreezeInitiatedBy = req.InitiatedBy
	w.FrozenAt = req.FrozenAt
	w.UpdatedAt = time.Now()

	return nil
}

// Unfreeze removes all transaction restrictions from a wallet
func (w *Wallet) Unfreeze() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if !w.Frozen {
		return ErrWalletNotFrozen
	}

	// Clear all freeze-related fields
	w.Frozen = false
	w.FreezeReason = ""
	w.FreezeInitiatedBy = ""
	w.FrozenAt = time.Time{}
	w.UpdatedAt = time.Now()

	return nil
}

// IsFrozen checks if wallet is currently frozen
func (w *Wallet) IsFrozen() bool {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	return w.Frozen
}

// FreezeInfo contains information about the freeze operation
type FreezeInfo struct {
	WalletID    string    `json:"walletId"`          // Frozen wallet ID
	IsFrozen    bool      `json:"isFrozen"`          // Frozen flag
	Reason      string    `json:"freezeReason"`      // Freeze reason
	InitiatedBy string    `json:"freezeInitiatedBy"` // Who initiated the freeze
	FrozenAt    time.Time `json:"frozenAt"`          // When the freeze was applied
}

// GetFreezeInfo returns current freeze status
func (w *Wallet) GetFreezeInfo() *FreezeInfo {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	return &FreezeInfo{
		WalletID:    w.ID,
		IsFrozen:    w.Frozen,
		Reason:      w.FreezeReason,
		InitiatedBy: w.FreezeInitiatedBy,
		FrozenAt:    w.FrozenAt,
	}
}
