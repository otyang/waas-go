package account

import (
	"context"

	"github.com/otyang/waas-go"
	"github.com/uptrace/bun"
)

func (a *Account) Credit(ctx context.Context, params waas.CreditWalletParams) (*waas.CreditWalletResponse, error) {
	err := a.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		return nil
	})

	return nil, err
}

func (a *Account) Debit(ctx context.Context, params waas.DebitWalletParams) (*waas.DebitWalletResponse, error) {
	return nil, nil
}

func (a *Account) Swap(ctx context.Context, params waas.SwapRequestParams) (*waas.SwapWalletResponse, error) {
	return nil, nil
}

func (a *Account) Transfer(ctx context.Context, params waas.TransferRequestParams) (*waas.TransferResponse, error) {
	return nil, nil
}

func (a *Account) Reverse(ctx context.Context, transactionID string) (*waas.ReverseResponse, error) {
	return nil, nil
}
