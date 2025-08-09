package types

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// TransferTransaction contains details for transferring between wallets
type TransferRequest struct {
	Amount                decimal.Decimal     `json:"amount"`
	Fee                   decimal.Decimal     `json:"fee"`
	Description           string              `json:"description"`
	InitiatorID           string              `json:"initiatorId"`
	ExternalTransactionID string              `json:"externalTransactionID"`
	TransactionCategory   TransactionCategory `json:"transactionCategory"`
}

// Transfer moves funds from this wallet to a destination wallet.
// It performs full validation, executes the transfer atomically, and returns
// transaction history records for both wallets.
//
// Parameters:
//   - dest: The destination wallet receiving funds
//   - req: Transfer request details
//
// Returns:
//   - sourceHistory: Debit record for this wallet
//   - destHistory: Credit record for destination wallet
//   - error: Validation or processing error
func (w *Wallet) Transfer(dest *Wallet, req TransferRequest) (*TransactionHistory, *TransactionHistory, error) {
	// Lock both wallets for atomic operation
	w.mutex.Lock()
	dest.mutex.Lock()
	defer w.mutex.Unlock()
	defer dest.mutex.Unlock()

	// Validate currencies match
	if w.CurrencyCode != dest.CurrencyCode {
		return nil, nil, ErrCurrencyMismatch
	}

	// Prepare transaction histories
	now := time.Now()
	sourceHistory := &TransactionHistory{
		ID:                uuid.New().String(),
		WalletID:          w.ID,
		CurrencyCode:      w.CurrencyCode,
		InitiatorID:       req.InitiatorID,
		ExternalReference: req.ExternalTransactionID,
		Category:          req.TransactionCategory,
		Description:       req.Description,
		Amount:            req.Amount,
		Fee:               req.Fee,
		Type:              TypeDebit,
		BalanceBefore:     w.AvailableBalance,
		InitiatedAt:       now,
		Status:            StatusPending,
	}

	destHistory := &TransactionHistory{
		ID:                uuid.New().String(),
		WalletID:          dest.ID,
		CurrencyCode:      dest.CurrencyCode,
		InitiatorID:       req.InitiatorID,
		ExternalReference: req.ExternalTransactionID,
		Category:          req.TransactionCategory,
		Description:       req.Description,
		Amount:            req.Amount,
		Fee:               decimal.Zero, // Fees only apply to source
		Type:              TypeCredit,
		BalanceBefore:     dest.AvailableBalance,
		InitiatedAt:       now,
		Status:            StatusPending,
	}

	// Validate transfer
	if err := validateTransfer(w, dest, req); err != nil {
		markFailed(sourceHistory, destHistory)
		return sourceHistory, destHistory, err
	}

	// Execute transfer
	totalDebit := req.Amount.Add(req.Fee)
	w.AvailableBalance = w.AvailableBalance.Sub(totalDebit)
	dest.AvailableBalance = dest.AvailableBalance.Add(req.Amount)

	// Finalize histories
	completeTransfer(sourceHistory, destHistory, w.AvailableBalance, dest.AvailableBalance)

	return sourceHistory, destHistory, nil
}

// validateTransfer checks all transfer requirements
func validateTransfer(source, dest *Wallet, req TransferRequest) error {
	if err := source.CanBeDebited(); err != nil {
		return err
	}
	if err := dest.CanBeCredited(); err != nil {
		return err
	}
	if req.Amount.LessThanOrEqual(decimal.Zero) {
		return ErrInvalidAmount
	}
	if req.Fee.LessThan(decimal.Zero) {
		return ErrInvalidFee
	}
	if source.AvailableBalance.LessThan(req.Amount.Add(req.Fee)) {
		return ErrInsufficientFunds
	}
	return nil
}

// markFailed updates transaction histories to failed status
func markFailed(source, dest *TransactionHistory) {
	now := time.Now()
	source.Status = StatusFailed
	source.CompletedAt = now
	dest.Status = StatusFailed
	dest.CompletedAt = now
}

// completeTransfer finalizes successful transactions
func completeTransfer(source, dest *TransactionHistory, sourceBal, destBal decimal.Decimal) {
	now := time.Now()
	source.BalanceAfter = sourceBal
	source.Status = StatusCompleted
	source.CompletedAt = now

	dest.BalanceAfter = destBal
	dest.Status = StatusCompleted
	dest.CompletedAt = now
}

// ========================================
// =============== swap

// SwapRequest represents a currency swap operation between wallets
type SwapRequest struct {
	// SourceAmount is the amount being sent from the source wallet (must be positive)
	SourceAmount decimal.Decimal `json:"sourceAmount"`

	// DestinationAmount is the amount being received in the destination wallet (must be positive)
	DestinationAmount decimal.Decimal `json:"destinationAmount"`

	// ExchangeRate is the rate used for the currency conversion
	ExchangeRate decimal.Decimal `json:"exchangeRate"`

	// Fee is the transaction fee being charged (must be zero or positive)
	Fee decimal.Decimal `json:"fee"`

	// Description provides context for the transaction
	Description string `json:"description"`

	// InitiatorID identifies who initiated the transaction
	InitiatorID string `json:"initiatorId"`

	// ExternalTransactionID is a reference ID from an external system
	ExternalTransactionID string `json:"externalTransactionID"`

	// TransactionCategory classifies the type of transaction
	TransactionCategory TransactionCategory `json:"transactionCategory"`
}

// Swap exchanges funds between wallets of different currencies at a specified rate.
// It validates the exchange rate, executes the swap atomically, and returns
// transaction history records for both wallets.
//
// Parameters:
//   - dest: The destination wallet receiving funds
//   - req: Swap request details including exchange rate
//
// Returns:
//   - sourceHistory: Debit record for this wallet (source currency)
//   - destHistory: Credit record for destination wallet (target currency)
//   - error: Validation or processing error
func (w *Wallet) Swap(dest *Wallet, req SwapRequest) (*TransactionHistory, *TransactionHistory, error) {
	// Lock both wallets for atomic operation
	w.mutex.Lock()
	dest.mutex.Lock()
	defer w.mutex.Unlock()
	defer dest.mutex.Unlock()

	// Prepare transaction histories
	now := time.Now()
	sourceHistory := &TransactionHistory{
		ID:                uuid.New().String(),
		WalletID:          w.ID,
		CurrencyCode:      w.CurrencyCode,
		InitiatorID:       req.InitiatorID,
		ExternalReference: req.ExternalTransactionID,
		Category:          req.TransactionCategory,
		Description:       req.Description,
		Amount:            req.SourceAmount,
		Fee:               req.Fee,
		Type:              TypeDebit,
		BalanceBefore:     w.AvailableBalance,
		InitiatedAt:       now,
		Status:            StatusPending,
	}

	destHistory := &TransactionHistory{
		ID:                uuid.New().String(),
		WalletID:          dest.ID,
		CurrencyCode:      dest.CurrencyCode,
		InitiatorID:       req.InitiatorID,
		ExternalReference: req.ExternalTransactionID,
		Category:          req.TransactionCategory,
		Description:       req.Description,
		Amount:            req.DestinationAmount,
		Fee:               decimal.Zero, // Fees only apply to source
		Type:              TypeCredit,
		BalanceBefore:     dest.AvailableBalance,
		InitiatedAt:       now,
		Status:            StatusPending,
	}

	// Validate swap
	if err := validateSwap(w, dest, req); err != nil {
		markFailed(sourceHistory, destHistory)
		return sourceHistory, destHistory, err
	}

	// Execute swap
	totalDebit := req.SourceAmount.Add(req.Fee)
	w.AvailableBalance = w.AvailableBalance.Sub(totalDebit)
	dest.AvailableBalance = dest.AvailableBalance.Add(req.DestinationAmount)

	// Finalize histories
	completeTransfer(sourceHistory, destHistory, w.AvailableBalance, dest.AvailableBalance)

	return sourceHistory, destHistory, nil
}

// validateSwap checks all swap requirements including exchange rate
func validateSwap(source, dest *Wallet, req SwapRequest) error {
	// Standard transfer validations
	if err := source.CanBeDebited(); err != nil {
		return err
	}
	if err := dest.CanBeCredited(); err != nil {
		return err
	}
	if req.SourceAmount.LessThanOrEqual(decimal.Zero) {
		return ErrInvalidAmount
	}
	if req.DestinationAmount.LessThanOrEqual(decimal.Zero) {
		return ErrInvalidAmount
	}
	if req.Fee.LessThan(decimal.Zero) {
		return ErrInvalidFee
	}
	if source.AvailableBalance.LessThan(req.SourceAmount.Add(req.Fee)) {
		return ErrInsufficientFunds
	}

	// Swap-specific validations
	if req.ExchangeRate.LessThanOrEqual(decimal.Zero) {
		return ErrInvalidExchangeRate
	}
	if !req.SourceAmount.Mul(req.ExchangeRate).Equal(req.DestinationAmount) {
		return ErrExchangeRateMismatch
	}

	return nil
}
