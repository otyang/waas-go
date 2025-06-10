package account

import (
	"context"
	"errors"

	"github.com/otyang/waas-go/types"
	"github.com/shopspring/decimal"
)

type (
	CreditOrDebitOption struct {
		WalletID               string
		Amount                 decimal.Decimal
		Fee                    decimal.Decimal
		PendTransaction        bool
		TxnCategory            types.TransactionCategory
		Status                 types.TransactionStatus
		Narration              *string `json:"narration"`
		UseThisAsTransactionID string
	}
	CreditOrDebitWalletResponse struct {
		Wallet      *types.Wallet
		Transaction *types.Transaction
	}
)

func (x *CreditOrDebitOption) Validate() error {
	if x.TxnCategory == "" {
		return errors.New("transaction category shouldn't be empty")
	}

	if x.Narration == nil {
		return errors.New("transaction narration shouldn't be empty")
	}

	if x.PendTransaction {
		x.Status = types.TransactionStatusPending
	}

	return nil
}

func (a *Client) CreditWallet(ctx context.Context, opts CreditOrDebitOption) (*CreditOrDebitWalletResponse, error) {
	wallet, err := a.FindWalletByID(ctx, opts.WalletID)
	if err != nil {
		return nil, err
	}

	t, w, err := types.CreditBalanceWithTxn(wallet, types.CreditOrDebitWalletOption{
		Amount:                         opts.Amount,
		Fee:                            opts.Fee,
		PendTransaction:                opts.PendTransaction,
		TxnCategory:                    opts.TxnCategory,
		Status:                         opts.Status,
		Narration:                      opts.Narration,
		OptionalUseThisAsTransactionID: opts.UseThisAsTransactionID,
	})
	if err != nil {
		return nil, err
	}

	err = a.WithTxBulkUpdateWalletAndInsertTransaction(ctx, []*types.Wallet{w}, []*types.Transaction{t})
	if err != nil {
		return nil, err
	}

	return &CreditOrDebitWalletResponse{Wallet: w, Transaction: t}, nil
}

func (a *Client) DebitWallet(ctx context.Context, opts CreditOrDebitOption) (*CreditOrDebitWalletResponse, error) {
	wallet, err := a.FindWalletByID(ctx, opts.WalletID)
	if err != nil {
		return nil, err
	}

	t, w, err := types.DebitBalanceWithTxn(wallet, types.CreditOrDebitWalletOption{
		Amount:                         opts.Amount,
		Fee:                            opts.Fee,
		PendTransaction:                opts.PendTransaction,
		TxnCategory:                    opts.TxnCategory,
		Status:                         opts.Status,
		Narration:                      opts.Narration,
		OptionalUseThisAsTransactionID: opts.UseThisAsTransactionID,
	})
	if err != nil {
		return nil, err
	}

	err = a.WithTxBulkUpdateWalletAndInsertTransaction(ctx, []*types.Wallet{w}, []*types.Transaction{t})
	if err != nil {
		return nil, err
	}

	return &CreditOrDebitWalletResponse{Wallet: w, Transaction: t}, nil
}

type (
	// TransferOpts defines parameters for transferring funds between wallets.
	TransferOption struct {
		FromWalletID    string          `json:"fromWid"`
		ToWalletID      string          `json:"toWid"`
		Amount          decimal.Decimal `json:"amount"`
		Fee             decimal.Decimal `json:"fee"`
		Narration       string          `json:"narration"`
		PendTransaction bool
		TxnCategory     types.TransactionCategory
		Status          types.TransactionStatus
	}

	// SwapOpts defines parameters for swapping currencies between wallets.
	SwapOption struct {
		CustomerID       string
		FromCurrencyCode string
		ToCurrencyCode   string
		FromAmount       decimal.Decimal
		FromFee          decimal.Decimal
		ToAmount         decimal.Decimal
		PendTransaction  bool
		TxnCategory      types.TransactionCategory
		Status           types.TransactionStatus
		Narration        *string `json:"narration"`
	}

	TransferOrSwapResponse struct {
		FromWallet      *types.Wallet
		ToWallet        *types.Wallet
		FromTransaction *types.Transaction
		ToTransaction   *types.Transaction
	}
)

func (a *Client) Transfer(ctx context.Context, opts TransferOption) (*TransferOrSwapResponse, error) {
	fromWallet, err := a.FindWalletByID(ctx, opts.FromWalletID)
	if err != nil {
		return nil, err
	}

	toWallet, err := a.FindWalletByID(ctx, opts.ToWalletID)
	if err != nil {
		return nil, err
	}

	fromTXN, toTXN, err := types.TransferWithTxn(fromWallet, toWallet, types.TransferWalletOption{
		Amount:          opts.Amount,
		Fee:             opts.Fee,
		PendTransaction: opts.PendTransaction,
		TxnCategory:     opts.TxnCategory,
		Status:          opts.Status,
		Narration:       &opts.Narration,
	})
	if err != nil {
		return nil, err
	}

	err = a.WithTxBulkUpdateWalletAndInsertTransaction(
		ctx, []*types.Wallet{fromWallet, toWallet}, []*types.Transaction{fromTXN, toTXN},
	)

	return &TransferOrSwapResponse{fromWallet, toWallet, fromTXN, toTXN}, err
}

func (a *Client) Swap(ctx context.Context, opts SwapOption) (*TransferOrSwapResponse, error) {
	fromWallet, err := a.FindWalletByCurrencyCode(ctx, opts.FromCurrencyCode, opts.CustomerID)
	if err != nil {
		return nil, err
	}

	toWallet, err := a.FindWalletByCurrencyCode(ctx, opts.ToCurrencyCode, opts.CustomerID)
	if err != nil {
		return nil, err
	}

	fromTXN, toTXN, err := types.SwapWithTxn(fromWallet, toWallet, types.SwapWalletOption{
		FromAmount:      opts.FromAmount,
		ToAmount:        opts.ToAmount,
		Fee:             opts.FromFee,
		PendTransaction: opts.PendTransaction,
		TxnCategory:     opts.TxnCategory,
		Status:          opts.Status,
		Narration:       opts.Narration,
	})
	if err != nil {
		return nil, err
	}

	err = a.WithTxBulkUpdateWalletAndInsertTransaction(
		ctx, []*types.Wallet{fromWallet, toWallet}, []*types.Transaction{fromTXN, toTXN},
	)

	return &TransferOrSwapResponse{fromWallet, toWallet, fromTXN, toTXN}, err
}
