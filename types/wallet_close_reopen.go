package types

import (
	"time"

	"github.com/google/uuid"
)

// CloseOrOpenRequest contains common fields for wallet state change operations
type CloseOrOpenRequest struct {
	RequestID   string    `json:"requestId"`   // Unique operation ID
	RequestedAt time.Time `json:"requestedAt"` // Timestamp of request
	Reason      string    `json:"reason"`      // Reason for operation
	InitiatedBy string    `json:"initiatedBy"` // Who initiated the operation
}

// CloseOrOpenResult contains common fields for wallet state change results
type CloseOrOpenResult struct {
	WalletID    string    `json:"walletId"`    // Affected wallet ID
	OperationID string    `json:"operationId"` // Unique ID for this operation
	ExecutedAt  time.Time `json:"executedAt"`  // When operation completed
	Balance     string    `json:"balance"`     // Balance snapshot
	Reason      string    `json:"reason"`      // Operation reason
}

// CloseWallet closes the wallet if it has zero balance
func (w *Wallet) CloseWallet(req CloseOrOpenRequest) (*CloseOrOpenResult, error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.IsClosed {
		return nil, ErrWalletAlreadyClosed
	}
	if !w.AvailableBalance.IsZero() || !w.LienBalance.IsZero() {
		return nil, ErrWalletNotEmpty
	}

	w.IsClosed = true
	w.UpdatedAt = time.Now()

	return &CloseOrOpenResult{
		WalletID:    w.ID,
		OperationID: uuid.New().String(),
		ExecutedAt:  time.Now(),
		Balance:     w.TotalBalance().String(),
		Reason:      req.Reason,
	}, nil
}

// ReopenWallet reopens a closed wallet
func (w *Wallet) ReopenWallet(req CloseOrOpenRequest) (*CloseOrOpenResult, error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if !w.IsClosed {
		return nil, ErrWalletNotClosed
	}
	if w.Frozen {
		return nil, ErrWalletFrozen
	}

	w.IsClosed = false
	w.UpdatedAt = time.Now()

	return &CloseOrOpenResult{
		WalletID:    w.ID,
		OperationID: uuid.New().String(),
		ExecutedAt:  time.Now(),
		Balance:     w.TotalBalance().String(),
		Reason:      req.Reason,
	}, nil
}
