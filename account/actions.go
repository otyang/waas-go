package account

import (
	"context"

	"github.com/otyang/waas-go/types"
	"github.com/uptrace/bun"
)

func (a *Account) WithTxBulkUpdateWalletAndInsertTransaction(
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

func (a *Account) Credit(ctx context.Context, params types.CreditWalletParams) (*types.CreditWalletResponse, error) {
	wallet, err := a.GetWalletByID(ctx, params.WalletID)
	if err != nil {
		return nil, err
	}

	err = wallet.CreditBalance(params.Amount, params.Fee)
	if err != nil {
		return nil, err
	}

	transaction := types.NewTransactionForCreditEntry(wallet, params.Amount, params.Fee, params.Type)
	transaction.SetNarration(params.Narration)

	err = a.WithTxBulkUpdateWalletAndInsertTransaction(ctx, []*types.Wallet{wallet}, []*types.Transaction{transaction})
	if err != nil {
		return nil, err
	}

	return &types.CreditWalletResponse{Wallet: wallet, Transaction: transaction}, nil
}

func (a *Account) Debit(ctx context.Context, params types.DebitWalletParams) (*types.DebitWalletResponse, error) {
	wallet, err := a.GetWalletByID(ctx, params.WalletID)
	if err != nil {
		return nil, err
	}

	err = wallet.DebitBalance(params.Amount, params.Fee)
	if err != nil {
		return nil, err
	}

	transaction := types.NewTransactionForDebitEntry(wallet, params.Amount, params.Fee, params.Type, params.Status)
	transaction.SetNarration(params.Narration)

	err = a.WithTxBulkUpdateWalletAndInsertTransaction(ctx, []*types.Wallet{wallet}, []*types.Transaction{transaction})
	if err != nil {
		return nil, err
	}

	return &types.DebitWalletResponse{Wallet: wallet, Transaction: transaction}, nil
}

func (a *Account) Transfer(ctx context.Context, params types.TransferRequestParams) (*types.TransferResponse, error) {
	fromWallet, err := a.GetWalletByID(ctx, params.FromWalletID)
	if err != nil {
		return nil, err
	}

	toWallet, err := a.GetWalletByID(ctx, params.ToWalletID)
	if err != nil {
		return nil, err
	}

	err = fromWallet.TransferTo(toWallet, params.Amount, params.Fee)
	if err != nil {
		return nil, err
	}

	fromTrsn, toTrsn := types.NewTransactionForTransfer(fromWallet, toWallet, params.Amount, params.Fee)
	fromTrsn.SetNarration(params.Narration)
	toTrsn.SetNarration(params.Narration)

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

func (a *Account) Swap(ctx context.Context, params types.SwapRequestParams) (*types.SwapWalletResponse, error) {
	fromWallet, err := a.GetWalletByUserIDAndCurrencyCode(ctx, params.CustomerID, params.FromCurrencyCode)
	if err != nil {
		return nil, err
	}

	toWallet, err := a.GetWalletByUserIDAndCurrencyCode(ctx, params.CustomerID, params.ToCurrencyCode)
	if err != nil {
		return nil, err
	}

	err = fromWallet.Swap(toWallet, params.FromAmount, params.ToAmount, params.FromFee)
	if err != nil {
		return nil, err
	}

	fromTrsn, toTrsn := types.NewTransactionForSwap(fromWallet, toWallet, params.FromAmount, params.ToAmount, params.FromFee)

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

func (a *Account) Reverse(ctx context.Context, transactionID string) (*types.ReverseResponse, error) {
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
