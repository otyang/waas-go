package types

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// CreditTransaction contains details for crediting a wallet
type CreditTransaction struct {
	Amount                decimal.Decimal     `json:"amount"`                // Positive amount to credit
	Fee                   decimal.Decimal     `json:"fee"`                   // Non-negative fee amount
	Description           string              `json:"description"`           // Human-readable context
	InitiatorID           string              `json:"initiatorId"`           // Who initiated the action
	ExternalTransactionID string              `json:"externalTransactionID"` // External system reference
	TransactionCategory   TransactionCategory `json:"transactionCategory"`   // Transaction classification
}

// DebitTransaction contains details for debiting a wallet
type DebitTransaction struct {
	Amount                decimal.Decimal     `json:"amount"`
	Fee                   decimal.Decimal     `json:"fee"`
	Description           string              `json:"description"`
	InitiatorID           string              `json:"initiatorId"`
	ExternalTransactionID string              `json:"externalTransactionID"`
	TransactionCategory   TransactionCategory `json:"transactionCategory"`
}

// TransactionRecord provides a complete audit record of a transaction
type TransactionRecord struct {
	TransactionID          string              `json:"transactionId"`          // Unique identifier
	WalletID               string              `json:"walletId"`               // Affected wallet
	CurrencyCode           string              `json:"currencyCode"`           // Currency type
	InitiatorID            string              `json:"initiatorId"`            // Who initiated
	ExternalTransactionRef string              `json:"externalTransactionRef"` // External reference
	TransactionCategory    TransactionCategory `json:"transactionCategory"`    // Transaction type
	Description            string              `json:"description"`            // Human-readable context
	Amount                 decimal.Decimal     `json:"amount"`                 // Principal amount
	Fee                    decimal.Decimal     `json:"fee"`                    // Applied fee
	TransactionType        TransactionType     `json:"transactionType"`        // Credit/Debit
	BalanceBefore          decimal.Decimal     `json:"balanceBefore"`          // Pre-transaction balance
	BalanceAfter           decimal.Decimal     `json:"balanceAfter"`           // Post-transaction balance
	CreatedAt              time.Time           `json:"createdAt"`              // When initiated
	UpdatedAt              time.Time           `json:"updatedAt"`              // Last updated
	Status                 TransactionStatus   `json:"status"`                 // Current state
}

// Credit adds funds to the wallet and returns a detailed transaction record
func (w *Wallet) Credit(tx CreditTransaction) (*TransactionRecord, error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	// Prepare initial transaction result
	result := &TransactionRecord{
		TransactionID:          uuid.New().String(),
		WalletID:               w.ID,
		CurrencyCode:           w.CurrencyCode,
		InitiatorID:            tx.InitiatorID,
		ExternalTransactionRef: tx.ExternalTransactionID,
		TransactionCategory:    tx.TransactionCategory,
		Description:            tx.Description,
		Amount:                 tx.Amount,
		Fee:                    tx.Fee,
		TransactionType:        TypeCredit,
		BalanceBefore:          w.AvailableBalance,
		CreatedAt:              time.Now(),
		UpdatedAt:              time.Now(),
		Status:                 StatusPending,
	}

	// Validate wallet state
	if err := w.CanBeCredited(); err != nil {
		result.Status = StatusFailed
		result.UpdatedAt = time.Now()
		return result, err
	}

	// Validate transaction amounts
	if tx.Amount.LessThanOrEqual(decimal.Zero) {
		result.Status = StatusFailed
		result.UpdatedAt = time.Now()
		return result, ErrInvalidAmount
	}
	if tx.Fee.LessThan(decimal.Zero) {
		result.Status = StatusFailed
		result.UpdatedAt = time.Now()
		return result, ErrInvalidFee
	}

	// Calculate net amount (amount - fee)
	netAmount := tx.Amount.Sub(tx.Fee)

	// Handle the credit operation
	if netAmount.LessThan(decimal.Zero) {
		// When fee exceeds amount, credit full amount without deducting fee
		w.AvailableBalance = w.AvailableBalance.Add(tx.Amount)
	} else {
		// Standard case - apply net amount
		w.AvailableBalance = w.AvailableBalance.Add(netAmount)
	}

	// Finalize result
	result.BalanceAfter = w.AvailableBalance
	result.Status = StatusCompleted
	result.UpdatedAt = time.Now()

	return result, nil
}

// Debit removes funds from the wallet and returns a detailed transaction record
func (w *Wallet) Debit(tx DebitTransaction) (*TransactionRecord, error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	// Prepare initial transaction result
	result := &TransactionRecord{
		TransactionID:          uuid.New().String(),
		WalletID:               w.ID,
		CurrencyCode:           w.CurrencyCode,
		InitiatorID:            tx.InitiatorID,
		ExternalTransactionRef: tx.ExternalTransactionID,
		TransactionCategory:    tx.TransactionCategory,
		Description:            tx.Description,
		Amount:                 tx.Amount,
		Fee:                    tx.Fee,
		TransactionType:        TypeDebit,
		BalanceBefore:          w.AvailableBalance,
		CreatedAt:              time.Now(),
		UpdatedAt:              time.Now(),
		Status:                 StatusPending,
	}

	// Validate wallet state
	if err := w.CanBeDebited(); err != nil {
		result.Status = StatusFailed
		result.UpdatedAt = time.Now()
		return result, err
	}

	// Validate transaction amounts
	if tx.Amount.LessThanOrEqual(decimal.Zero) {
		result.Status = StatusFailed
		result.UpdatedAt = time.Now()
		return result, ErrInvalidAmount
	}
	if tx.Fee.LessThan(decimal.Zero) {
		result.Status = StatusFailed
		result.UpdatedAt = time.Now()
		return result, ErrInvalidFee
	}

	// Calculate total debit (amount + fee)
	totalDebit := tx.Amount.Add(tx.Fee)

	// Check sufficient funds
	if w.AvailableBalance.LessThan(totalDebit) {
		result.Status = StatusFailed
		result.UpdatedAt = time.Now()
		return result, ErrInsufficientFunds
	}

	// Perform the debit
	w.AvailableBalance = w.AvailableBalance.Sub(totalDebit)

	// Finalize result
	result.BalanceAfter = w.AvailableBalance
	result.Status = StatusCompleted
	result.UpdatedAt = time.Now()

	return result, nil
}
