package account

import (
	"context"
	"time"

	"github.com/otyang/waas-go/types"
	"github.com/uptrace/bun"
)

func (a *Account) CreateTransaction(ctx context.Context, transaction *types.Transaction) (*types.Transaction, error) {
	transaction.CreatedAt = time.Now()
	transaction.UpdatedAt = time.Now()
	_, err := a.db.NewInsert().Model(transaction).Exec(ctx)
	return transaction, err
}

func (a *Account) GetTransaction(ctx context.Context, transactionID string) (*types.Transaction, error) {
	transaction := types.Transaction{ID: transactionID}
	err := a.db.NewSelect().Model(&transaction).WherePK().Limit(1).Scan(ctx)
	return &transaction, err
}

func (a *Account) UpdateTransaction(ctx context.Context, transaction *types.Transaction) (*types.Transaction, error) {
	transaction.UpdatedAt = time.Now()
	_, err := a.db.NewUpdate().Model(transaction).WherePK().Exec(ctx)
	return transaction, err
}

func (a *Account) UpdateTransactionStatus(ctx context.Context, transactionID string, status types.TransactionStatus) (*types.Transaction, error) {
	transaction := &types.Transaction{ID: transactionID, Status: status, UpdatedAt: time.Now()}
	_, err := a.db.NewUpdate().Model(transaction).Column("status", "updated_at").WherePK().Exec(ctx)
	return transaction, err
}

func (a *Account) ListTransaction(ctx context.Context, params types.ListTransactionsFilterParams) ([]types.Transaction, string, error) {
	var transactions []types.Transaction

	if params.Limit < 1 {
		params.Limit = 1
	}

	if params.Limit > 500 {
		params.Limit = 500
	}

	q := a.db.NewSelect().Model(&transactions).Limit(params.Limit + 1)

	{ // filters
		if params.CustomerID != nil {
			q.Where("customer_id = ?", params.CustomerID)
		}

		if params.WalletID != nil {
			q.Where("wallet_id = ?", params.WalletID)
		}

		if len(params.Currency) > 0 {
			q.Where("lower(currency) IN (?)", bun.In(types.ToLowercaseSlice(params.Currency)))
		}

		if params.IsDebit != nil {
			q.Where("is_debit = ?", params.IsDebit)
		}

		if params.Type != nil {
			q.Where("lower(type) = lower(?)", params.Type)
		}

		if params.Status != nil {
			q.Where("lower(status) = lower(?)", params.Status)
		}

		if params.Reversed != nil && *params.Reversed {
			q.Where("reversed_at IS NOT NULL")
		}

		if params.EndDate.IsZero() && params.StartDate.IsZero() {
			q.OrderExpr("created_at DESC")
		}

		if params.EndDate.IsZero() && !params.StartDate.IsZero() {
			q.Where("created_at >= ?", params.StartDate).OrderExpr("created_at ASC")
		}

		if !params.EndDate.IsZero() && params.StartDate.IsZero() {
			q.Where("created_at <= ?", params.EndDate).OrderExpr("created_at DESC")
		}

		if !params.EndDate.IsZero() && !params.StartDate.IsZero() {
			q.Where("created_at >= ?", params.StartDate)
			q.Where("created_at <= ?", params.EndDate)
			q.OrderExpr("created_at ASC")
		}
	}

	if err := q.Scan(ctx); err != nil {
		return nil, "", err
	}

	var nextCursor string
	if len(transactions) > params.Limit {
		nextCursor = transactions[params.Limit].CreatedAt.GoString()
		transactions = transactions[:params.Limit-1]
	}

	return transactions, nextCursor, nil
}
