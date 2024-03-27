package waas

import (
	"strings"
	"sync"
	"time"

	"github.com/shopspring/decimal"
	"github.com/uptrace/bun"
)

var (
	ErrWalletFrozen                   = NewWaasError("account frozen")
	ErrWalletInsufficientBalance      = NewWaasError("insufficient balance")
	ErrWalletInvalid                  = NewWaasError("invalid wallet")
	ErrWalletInvalidAmount            = NewWaasError("invalid amount: cannot be zero or negative")
	ErrWalletInvalidTransferSameOwner = NewWaasError("cannot transfer funds to the same owner or wallet")
	ErrWalletSameCurrencySwap         = NewWaasError("cannot swap between same currency")
	ErrWalletSwapSameOwnerRequired    = NewWaasError("cannot swap between diffetent customers")
	ErrWalletClosed                   = NewWaasError("cannot operate on a closed wallet")
)

type WalletStatus string

const (
	WalletStatusActive WalletStatus = "active"
	WalletStatusFrozen WalletStatus = "frozen"
	WalletStatusClosed WalletStatus = "closed"
)

// GenerateWalletID generates a wallet ID
func GenerateWalletID(currencyCode, userID string) string {
	return strings.ToLower(strings.TrimSpace(currencyCode)) + "-" + userID
}

// Wallet represents a user's wallet for holding and managing funds.
//   - ID is the unique identifier for the wallet.
//   - CustomerID is the ID of the user who owns the wallet.
//   - CurrencyCode is the ISO 4217 code of the currency used by the wallet.
//   - AvailableBalance is the amount of currency readily available for use.
//   - LienBalance is the amount of currency currently locked and unavailable.
//   - Status indicates whether the wallet is frozen, active or closed.
//   - IsFiat indicates whether the currency is a fiat currency (e.g., USD, EUR) or not.===================\\\\\\\
//   - CreatedAt represents the timestamp when the wallet was created.
//   - UpdatedAt represents the timestamp of the last update to the wallet.
//   - VersionID is a unique identifier used to ensure that the same operation is not performed multiple times.
type Wallet struct {
	bun.BaseModel    `bun:"table:wallets"`
	mutex            sync.Mutex      `bun:"-"`
	ID               string          `json:"id" bun:"id,pk"`
	CustomerID       string          `json:"customerId" bun:",notnull"`
	CurrencyCode     string          `json:"currencyCode" bun:",notnull"`
	AvailableBalance decimal.Decimal `json:"availableBalance" bun:"type:decimal(24,8),notnull"`
	LienBalance      decimal.Decimal `json:"lienBalance" bun:"type:decimal(24,8),notnull"`
	Status           WalletStatus    `json:"status" bun:",notnull"`
	CreatedAt        time.Time       `json:"createdAt" bun:",notnull"`
	UpdatedAt        time.Time       `json:"updatedAt" bun:",notnull"`
	VersionId        string          `json:"-" bun:",notnull"`
}

// NewWallet creates a new Wallet instance.
func NewWallet(customerID, currencyCode string, isFiat bool) *Wallet {
	return &Wallet{
		mutex:            sync.Mutex{},
		ID:               GenerateWalletID(currencyCode, customerID),
		CustomerID:       customerID,
		CurrencyCode:     strings.ToUpper(currencyCode),
		AvailableBalance: decimal.Zero,
		Status:           WalletStatusActive,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		VersionId:        GenerateID(7),
	}
}

// TotalBalance Gets the total balance
func (w *Wallet) TotalBalance() decimal.Decimal {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	return w.AvailableBalance.Add(w.LienBalance)
}

// Freeze freezes the wallet.
func (w *Wallet) Freeze() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.Status == WalletStatusClosed {
		return ErrWalletClosed
	}

	w.Status = WalletStatusFrozen
	w.UpdatedAt = time.Now()
	return nil
}

// Unfreeze unfreezes the wallet, allowing transactions to resume.
func (w *Wallet) Unfreeze() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.Status == WalletStatusClosed {
		return ErrWalletClosed
	}

	w.Status = WalletStatusActive
	w.UpdatedAt = time.Now()
	return nil
}

// IsActive checks if the wallet is active
func (w *Wallet) IsActive() bool {
	return w.Status == WalletStatusActive
}

// IsClosed checks if the wallet is closed.
func (w *Wallet) IsClosed() bool {
	return w.Status == WalletStatusClosed
}

// IsFrozen checks if the wallet is frozen.
func (w *Wallet) IsFrozen() bool {
	return w.Status == WalletStatusFrozen
}

// CreditBalance adds the specified amount to the available balance after subtracting the fee.
func (w *Wallet) CreditBalance(amount, fee decimal.Decimal) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.IsClosed() {
		return ErrWalletClosed
	}

	if amount.LessThanOrEqual(decimal.Zero) || fee.LessThan(decimal.Zero) {
		return ErrWalletInvalidAmount
	}

	// Avoid negative balances by bypassing fees if insufficient funds
	if w.AvailableBalance.Add(amount).Sub(fee).LessThan(decimal.Zero) {
		w.AvailableBalance = w.AvailableBalance.Add(amount)
	} else {
		w.AvailableBalance = w.AvailableBalance.Add(amount).Sub(fee)
	}

	return nil
}

// DebitBalance subtracts a specified amount from the AvailableBalance.
func (w *Wallet) DebitBalance(amount, fee decimal.Decimal) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.IsFrozen() {
		return ErrWalletFrozen
	}

	if w.Status == WalletStatusClosed {
		return ErrWalletClosed
	}

	if amount.LessThanOrEqual(decimal.Zero) || fee.LessThan(decimal.Zero) {
		return ErrWalletInvalidAmount
	}

	if w.AvailableBalance.Sub(amount).Sub(fee).LessThan(decimal.Zero) {
		return ErrWalletInsufficientBalance
	}

	w.AvailableBalance = w.AvailableBalance.Sub(amount).Sub(fee)

	return nil
}

// Transfer transfers a specified amount between wallets, taking into account fees.
func (fromWallet *Wallet) TransferTo(toWallet *Wallet, amount, fee decimal.Decimal) error {
	if fromWallet == nil || toWallet == nil {
		return ErrWalletInvalid
	}

	// ensure we not transferring to same wallet
	if fromWallet.ID == toWallet.ID {
		return ErrWalletInvalidTransferSameOwner
	}

	if fromWallet.IsClosed() || toWallet.IsClosed() {
		return ErrWalletClosed
	}

	err := fromWallet.DebitBalance(amount, fee)
	if err != nil {
		return err
	}

	err = toWallet.CreditBalance(amount, decimal.Zero)
	if err != nil {
		// Rollback transaction if credit fails (direct addition)
		fromWallet.AvailableBalance.Add(amount).Add(fee)
		return err
	}

	return nil
}

// Swap performs a swap between two wallets for different currencies.
func (fromWallet *Wallet) Swap(toWallet *Wallet, fromAmount, toAmount, fee decimal.Decimal) error {
	if fromWallet == nil || toWallet == nil {
		return ErrWalletInvalid
	}

	// Check if currencies are different
	if fromWallet.CurrencyCode == toWallet.CurrencyCode {
		return ErrWalletSameCurrencySwap
	}

	// Check to ensure swap action is to same user/owner
	if fromWallet.CustomerID != toWallet.CustomerID {
		return ErrWalletSwapSameOwnerRequired
	}

	if fromWallet.IsClosed() || toWallet.IsClosed() {
		return ErrWalletClosed
	}

	// Check if sufficient balance for swap
	err := fromWallet.DebitBalance(fromAmount, fee)
	if err != nil {
		return err
	}

	// Perform swap
	err = toWallet.CreditBalance(toAmount, decimal.Zero)
	if err != nil {
		// Add back deducted amount and fee in case of failure
		fromWallet.AvailableBalance.Add(fromAmount).Add(fee)
		return err
	}

	return nil
}

// LienAmount locks a specified amount from the available balance.
func (w *Wallet) LienAmount(amount decimal.Decimal) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if amount.LessThanOrEqual(decimal.Zero) {
		return ErrWalletInvalidAmount
	}

	if w.AvailableBalance.Sub(amount).LessThan(decimal.Zero) {
		return ErrWalletInsufficientBalance
	}

	w.AvailableBalance = w.AvailableBalance.Sub(amount)
	w.LienBalance = w.LienBalance.Add(amount)

	return nil
}

// UnLienAmount unlocks a specified amount from the lien balance.
func (w *Wallet) UnLienAmount(amount decimal.Decimal) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if amount.LessThanOrEqual(decimal.Zero) {
		return ErrWalletInvalidAmount
	}

	if w.LienBalance.Sub(amount).LessThan(decimal.Zero) {
		return ErrWalletInsufficientBalance
	}

	w.LienBalance = w.LienBalance.Sub(amount)
	w.AvailableBalance = w.AvailableBalance.Add(amount)

	return nil
}
