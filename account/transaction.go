package account

import (
	"context"

	"github.com/otyang/waas-go"
)

func (a *Account) CreateTransaction(ctx context.Context, transaction *waas.Transaction) (*waas.Transaction, error) {
	return nil, nil
}

func (a *Account) GetTransaction(ctx context.Context, transactionID string) (*waas.Transaction, error) {
	return nil, nil
}

func (a *Account) UpdateTransaction(ctx context.Context, transaction *waas.Transaction) (*waas.Transaction, error) {
	return nil, nil
}

func (a *Account) ListTransaction(ctx context.Context, params waas.ListTransactionsFilterParams) ([]waas.Transaction, error) {
	return nil, nil
}
