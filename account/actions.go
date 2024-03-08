package account

import (
	"context"

	"github.com/otyang/waas-go"
)

// decimal places issues
// validation of params

func (a *Account) Credit(ctx context.Context, params waas.CreditWalletParams) (*waas.CreditWalletResponse, error) {
	wallet, err := a.GetWalletByID(ctx, params.WalletID)
	if err != nil {
		return nil, err
	}

	err = wallet.CreditBalance(params.Amount, params.Fee)
	if err != nil {
		return nil, err
	}

	transaction := waas.NewTransactionForCreditEntry(wallet, params.Amount, params.Fee, params.Type)

	err = a.WithTxUpdateWalletAndTransaction(ctx, wallet, transaction)
	if err != nil {
		return nil, err
	}

	return &waas.CreditWalletResponse{Wallet: wallet, Transaction: transaction}, nil
}

func (a *Account) Debit(ctx context.Context, params waas.DebitWalletParams) (*waas.DebitWalletResponse, error) {
	wallet, err := a.GetWalletByID(ctx, params.WalletID)
	if err != nil {
		return nil, err
	}

	err = wallet.DebitBalance(params.Amount, params.Fee)
	if err != nil {
		return nil, err
	}

	transaction := waas.NewTransactionForDebitEntry(wallet, params.Amount, params.Fee, params.Type, params.Status)

	err = a.WithTxUpdateWalletAndTransaction(ctx, wallet, transaction)
	if err != nil {
		return nil, err
	}

	return &waas.DebitWalletResponse{Wallet: wallet, Transaction: transaction}, nil
}

func (a *Account) Transfer(ctx context.Context, params waas.TransferRequestParams) (*waas.TransferResponse, error) {
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

	fromTrsn, toTrsn := waas.NewTransactionForTransfer(fromWallet, toWallet, params.Amount, params.Fee)

	err = a.WithTxBulkUpdateWalletAndTransaction(ctx, []*waas.Wallet{fromWallet, toWallet}, []*waas.Transaction{fromTrsn, toTrsn})
	if err != nil {
		return nil, err
	}

	return &waas.TransferResponse{
		FromWallet:      fromWallet,
		ToWallet:        toWallet,
		FromTransaction: fromTrsn,
		ToTransaction:   toTrsn,
	}, nil
}

func (a *Account) Swap(ctx context.Context, params waas.SwapRequestParams) (*waas.SwapWalletResponse, error) {
	fromWallet, err := a.GetWalletByUserIDAndCurrencyCode(ctx, params.UserID, params.FromCurrencyCode)
	if err != nil {
		return nil, err
	}

	toWallet, err := a.GetWalletByUserIDAndCurrencyCode(ctx, params.UserID, params.ToCurrencyCode)
	if err != nil {
		return nil, err
	}

	err = fromWallet.Swap(toWallet, params.FromAmount, params.ToAmount, params.FromFee)
	if err != nil {
		return nil, err
	}

	fromTrsn, toTrsn := waas.NewTransactionForSwap(fromWallet, toWallet, params.FromAmount, params.ToAmount, params.FromFee)

	err = a.WithTxBulkUpdateWalletAndTransaction(ctx, []*waas.Wallet{fromWallet, toWallet}, []*waas.Transaction{fromTrsn, toTrsn})
	if err != nil {
		return nil, err
	}

	return &waas.SwapWalletResponse{
		FromWallet:      fromWallet,
		ToWallet:        toWallet,
		FromTransaction: fromTrsn,
		ToTransaction:   toTrsn,
	}, err
}

func (a *Account) Reverse(ctx context.Context, transactionID string) (*waas.ReverseResponse, error) {
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

	err = a.WithTxBulkUpdateWalletAndTransaction(ctx, []*waas.Wallet{rr.Wallet}, []*waas.Transaction{rr.OldTx, rr.NewTx})
	if err != nil {
		return nil, err
	}

	return rr, nil
}
