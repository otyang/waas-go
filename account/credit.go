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

	CreditWalletResponse struct {
		Wallet      *types.Wallet
		Transaction *types.Transaction
	}

	DebitWalletOption struct {
		WalletID        string
		Amount          decimal.Decimal
		Fee             decimal.Decimal
		PendTransaction bool
		Transaction     *types.Transaction
	}

	DebitWalletResponse struct {
		Wallet      *types.Wallet
		Transaction *types.Transaction
	}
)

func (x *CreditWalletOption) Validate() error {
	if x.Transaction == nil {
		return errors.New("transaction parameter shouldn't be empty")
	}

	if x.Transaction.Type == "" {
		return errors.New("transaction type parameter shouldn't be empty")
	}

	if x.Transaction.Narration == nil {
		return errors.New("transaction narration parameter shouldn't be empty")
	}

	if x.PendTransaction {
		x.Transaction.Status = types.TransactionStatusPending
	}

	x.Transaction.Amount = x.Amount
	x.Transaction.Fee = x.Fee
	x.Transaction.Total = x.Amount.Add(x.Fee)

	return nil
}

func (a *Client) CreditWallet(ctx context.Context, opts CreditWalletOption) (*CreditWalletResponse, error) {
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

	transaction := types.NewTransactionForCreditEntry(wallet, opts.Amount, opts.Fee, opts.Transaction.Type)

	err = a.WithTxBulkUpdateWalletAndInsertTransaction(ctx, []*types.Wallet{wallet}, []*types.Transaction{transaction})
	if err != nil {
		return nil, err
	}

	return &CreditWalletResponse{Wallet: wallet, Transaction: transaction}, nil
}

// ==============================

func (x *DebitWalletOption) Validate() error {
	if x.Transaction == nil {
		return errors.New("transaction parameter shouldn't be empty")
	}

	if x.Transaction.Type == "" {
		return errors.New("transaction type parameter shouldn't be empty")
	}

	if x.Transaction.Narration == nil {
		return errors.New("transaction narration parameter shouldn't be empty")
	}

	x.Transaction.Amount = x.Amount
	x.Transaction.Fee = x.Fee
	x.Transaction.Total = x.Amount.Add(x.Fee)

	return nil
}

func (a *Client) DebitWallet(ctx context.Context, opts CreditWalletOption) (*CreditWalletResponse, error) {
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

	transaction := types.NewTransactionForDebitEntry(wallet, opts.Amount, opts.Fee, opts.Transaction.Type)

	err = a.WithTxBulkUpdateWalletAndInsertTransaction(ctx, []*types.Wallet{wallet}, []*types.Transaction{transaction})
	if err != nil {
		return nil, err
	}

	return &CreditWalletResponse{Wallet: wallet, Transaction: transaction}, nil
}

// Creates a new credit transaction entry.
func NewTransactionForCreditEntry(wallet *Wallet, amount, fee decimal.Decimal, txn types.Transaction) *Transaction {
	return &types.Transaction{
		ID:             types.NewTransactionID(),
		CustomerID:     wallet.CustomerID,
		WalletID:       wallet.ID,
		IsDebit:        false,
		Currency:       wallet.CurrencyCode,
		Amount:         amount,
		Fee:            fee,
		BalanceAfter:   wallet.AvailableBalance,
		Type:           txn.Type,
		Status:         txn.Status,
		Narration:      nil,
		CounterpartyID: nil,
		ReversedAt:     nil,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}
