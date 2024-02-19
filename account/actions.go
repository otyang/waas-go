package account

import (
	"context"

	"github.com/otyang/waas-go"
)

func (a *Account) Credit(ctx context.Context, params waas.CreditWalletParams) (*waas.CreditWalletResponse, error) {
	return nil, nil
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
