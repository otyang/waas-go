package account

import (
	"context"
	"errors"
	"time"

	"github.com/otyang/waas-go/types"
)

type (
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

func (a *Client) CreditWallet(ctx context.Context, opts types.CreditOrDebitWalletOption) (*CreditOrDebitWalletResponse, error) {
	wallet, err := a.FindWalletByID(ctx, opts.WalletID)
	if err != nil {
		return nil, err
	}

	t, w, err := types.CreditBalanceWithTxn(wallet, opts)
	if err != nil {
		return nil, err
	}

	err = a.WithTxBulkUpdateWalletAndInsertTransaction(ctx, []*types.Wallet{w}, []*types.Transaction{t})
	if err != nil {
		return nil, err
	}

	return &CreditOrDebitWalletResponse{Wallet: w, Transaction: t}, nil
}

// ==============================

func (a *Client) DebitWallet(ctx context.Context, opts types.CreditOrDebitWalletOption) (*CreditOrDebitWalletResponse, error) {
	wallet, err := a.FindWalletByID(ctx, opts.WalletID)
	if err != nil {
		return nil, err
	}

	t, w, err := types.DebitBalanceWithTxn(wallet, opts)
	if err != nil {
		return nil, err
	}

	err = a.WithTxBulkUpdateWalletAndInsertTransaction(ctx, []*types.Wallet{w}, []*types.Transaction{t})
	if err != nil {
		return nil, err
	}

	return &CreditOrDebitWalletResponse{Wallet: w, Transaction: t}, nil
}
