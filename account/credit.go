package account

import (
	"context"
	"errors"
	"time"

	"github.com/otyang/waas-go/types"
	"github.com/shopspring/decimal"
)

type (
	CreditWalletOption struct {
		WalletID        string
		Amount          decimal.Decimal
		Fee             decimal.Decimal
		PendTransaction bool
		Transaction     *types.Transaction
	}

	DebitWalletOption struct {
		WalletID        string
		Amount          decimal.Decimal
		Fee             decimal.Decimal
		PendTransaction bool
		Transaction     *types.Transaction
	}

	CreditOrDebitWalletResponse struct {
		Wallet      *types.Wallet
		Transaction *types.Transaction
	}
)

func (x *CreditWalletOption) Validate() error {
	if x.Transaction == nil {
		return errors.New("transaction parameter shouldn't be empty")
	}

	if x.Transaction.Category == "" {
		return errors.New("transaction category parameter shouldn't be empty")
	}

	if x.Transaction.Narration == nil {
		return errors.New("transaction narration parameter shouldn't be empty")
	}

	if x.PendTransaction {
		x.Transaction.Status = types.TransactionStatusPending
	}

	if x.Transaction.CreatedAt.IsZero() {
		x.Transaction.CreatedAt = time.Now()
	}

	if x.Transaction.UpdatedAt.IsZero() {
		x.Transaction.UpdatedAt = time.Now()
	}

	x.Transaction.Amount = x.Amount
	x.Transaction.Fee = x.Fee
	x.Transaction.Total = x.Amount.Add(x.Fee)

	return nil
}

func (a *Client) CreditWallet(ctx context.Context, opts CreditWalletOption) (*CreditOrDebitWalletResponse, error) {
	wallet, err := a.FindWalletByID(ctx, opts.WalletID)
	if err != nil {
		return nil, err
	}

	if err := opts.Validate(); err != nil {
		return nil, err
	}

	err = wallet.CreditBalance(opts.Amount, opts.Fee)
	if err != nil {
		return nil, err
	}

	transaction := NewTransaction(wallet, opts.Transaction, true)

	err = a.WithTxBulkUpdateWalletAndInsertTransaction(ctx, []*types.Wallet{wallet}, []*types.Transaction{transaction})
	if err != nil {
		return nil, err
	}

	return &CreditOrDebitWalletResponse{Wallet: wallet, Transaction: transaction}, nil
}

// ==============================

func (a *Client) DebitWallet(ctx context.Context, opts CreditWalletOption) (*CreditOrDebitWalletResponse, error) {
	wallet, err := a.FindWalletByID(ctx, opts.WalletID)
	if err != nil {
		return nil, err
	}

	if err := opts.Validate(); err != nil {
		return nil, err
	}

	err = wallet.DebitBalance(opts.Amount, opts.Fee)
	if err != nil {
		return nil, err
	}

	transaction := NewTransaction(wallet, opts.Transaction, false)

	err = a.WithTxBulkUpdateWalletAndInsertTransaction(ctx, []*types.Wallet{wallet}, []*types.Transaction{transaction})
	if err != nil {
		return nil, err
	}

	return &CreditOrDebitWalletResponse{Wallet: wallet, Transaction: transaction}, nil
}

// Creates a new credit transaction entry.
func NewTransaction(wallet *types.Wallet, txn *types.Transaction, isCredit bool) *types.Transaction {
	return &types.Transaction{
		ID:           txn.ID,
		CustomerID:   wallet.CustomerID,
		WalletID:     wallet.ID,
		IsDebit:      !isCredit,
		Currency:     wallet.CurrencyCode,
		Amount:       txn.Amount,
		Fee:          txn.Fee,
		Total:        txn.Total,
		BalanceAfter: wallet.AvailableBalance,
		Type:         txn.Type,
		Status:       txn.Status,
		Narration:    txn.Narration,
		ServiceTxnID: txn.ServiceTxnID,
		ReversedAt:   nil,
		CreatedAt:    txn.CreatedAt,
		UpdatedAt:    txn.UpdatedAt,
	}
}
