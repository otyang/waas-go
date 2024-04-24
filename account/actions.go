package account

import (
	"context"

	"github.com/otyang/waas-go/types"
	"github.com/uptrace/bun"
)

func (a *Client) WithTxBulkUpdateWalletAndInsertTransaction(
	ctx context.Context,
	wallets []*types.Wallet,
	transactions []*types.Transaction,
	otherEntries ...any,
) error {
	return a.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		for _, wallet := range wallets {
			if wallet == nil {
				continue
			}
			_, err := a.NewWithTx(tx).UpdateWallet(ctx, wallet)
			if err != nil {
				return err
			}
		}

		if len(transactions) > 0 {
			for _, transaction := range transactions {
				if transaction == nil {
					continue
				}
				_, err := a.NewWithTx(tx).CreateTransaction(ctx, transaction)
				if err != nil {
					return err
				}
			}
		}

		// for other entries. transaction specific structs
		if len(otherEntries) > 0 {
			for _, otherEntry := range otherEntries {
				if otherEntry == nil {
					continue
				}
				_, err := tx.NewInsert().Model(otherEntry).Exec(ctx)
				if err != nil {
					return err
				}
			}
		}

		return nil
	})
}

func (a *Client) Credit(ctx context.Context, opts types.CreditWalletOpts) (*types.CreditWalletResponse, error) {
	wallet, err := a.GetWalletByID(ctx, opts.WalletID)
	if err != nil {
		return nil, err
	}

	err = wallet.CreditBalance(opts.Amount, opts.Fee)
	if err != nil {
		return nil, err
	}

	transaction := types.NewTransactionForCreditEntry(wallet, opts.Amount, opts.Fee, opts.Type)
	transaction.SetNarration(opts.Narration)

	err = a.WithTxBulkUpdateWalletAndInsertTransaction(ctx, []*types.Wallet{wallet}, []*types.Transaction{transaction})
	if err != nil {
		return nil, err
	}

	return &types.CreditWalletResponse{Wallet: wallet, Transaction: transaction}, nil
}

func (a *Client) Debit(ctx context.Context, opts types.DebitWalletOpts) (*types.DebitWalletResponse, error) {
	wallet, err := a.GetWalletByID(ctx, opts.WalletID)
	if err != nil {
		return nil, err
	}

	err = wallet.DebitBalance(opts.Amount, opts.Fee)
	if err != nil {
		return nil, err
	}

	transaction := types.NewTransactionForDebitEntry(wallet, opts.Amount, opts.Fee, opts.Type, opts.Status)
	transaction.SetNarration(opts.Narration)

	err = a.WithTxBulkUpdateWalletAndInsertTransaction(ctx, []*types.Wallet{wallet}, []*types.Transaction{transaction})
	if err != nil {
		return nil, err
	}

	return &types.DebitWalletResponse{Wallet: wallet, Transaction: transaction}, nil
}

func (a *Client) Transfer(ctx context.Context, opts types.TransferRequestOpts) (*types.TransferResponse, error) {
	fromWallet, err := a.GetWalletByID(ctx, opts.FromWalletID)
	if err != nil {
		return nil, err
	}

	toWallet, err := a.GetWalletByID(ctx, opts.ToWalletID)
	if err != nil {
		return nil, err
	}

	err = fromWallet.TransferTo(toWallet, opts.Amount, opts.Fee)
	if err != nil {
		return nil, err
	}

	fromTrsn, toTrsn := types.NewTransactionForTransfer(fromWallet, toWallet, opts.Amount, opts.Fee)
	fromTrsn.SetNarration(opts.Narration)
	toTrsn.SetNarration(opts.Narration)

	err = a.WithTxBulkUpdateWalletAndInsertTransaction(ctx, []*types.Wallet{fromWallet, toWallet}, []*types.Transaction{fromTrsn, toTrsn})
	if err != nil {
		return nil, err
	}

	return &types.TransferResponse{
		FromWallet:      fromWallet,
		ToWallet:        toWallet,
		FromTransaction: fromTrsn,
		ToTransaction:   toTrsn,
	}, nil
}

func (a *Client) Swap(ctx context.Context, opts types.SwapRequestOpts) (*types.SwapWalletResponse, error) {
	fromWallet, err := a.GetWalletByCurrencyCode(ctx, opts.CustomerID, opts.FromCurrencyCode)
	if err != nil {
		return nil, err
	}

	toWallet, err := a.GetWalletByCurrencyCode(ctx, opts.CustomerID, opts.ToCurrencyCode)
	if err != nil {
		return nil, err
	}

	err = fromWallet.Swap(toWallet, opts.FromAmount, opts.ToAmount, opts.FromFee)
	if err != nil {
		return nil, err
	}

	fromTrsn, toTrsn := types.NewTransactionForSwap(fromWallet, toWallet, opts.FromAmount, opts.ToAmount, opts.FromFee)

	err = a.WithTxBulkUpdateWalletAndInsertTransaction(ctx, []*types.Wallet{fromWallet, toWallet}, []*types.Transaction{fromTrsn, toTrsn})
	if err != nil {
		return nil, err
	}

	return &types.SwapWalletResponse{
		FromWallet:      fromWallet,
		ToWallet:        toWallet,
		FromTransaction: fromTrsn,
		ToTransaction:   toTrsn,
	}, err
}

func (a *Client) Reverse(ctx context.Context, transactionID string) (*types.ReverseResponse, error) {
	t, err := a.GetTransaction(ctx, transactionID)
	if err != nil {
		return nil, err
	}

	wallet, err := a.GetWalletByID(ctx, t.WalletID)
	if err != nil {
		return nil, err
	}

	rr, err := t.Reverse(wallet)
	if err != nil {
		return nil, err
	}
	// update transaction status
	err = a.WithTxBulkUpdateWalletAndInsertTransaction(ctx, []*types.Wallet{rr.Wallet}, []*types.Transaction{rr.OldTx, rr.NewTx})
	if err != nil {
		return nil, err
	}

	return rr, nil
}
