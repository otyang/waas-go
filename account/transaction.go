package account

import (
	"context"
	"strings"
	"time"

	"github.com/otyang/waas-go/types"
	"github.com/uptrace/bun"
)

func (a *Client) CreateTransaction(ctx context.Context, transaction *types.Transaction) (*types.Transaction, error) {
	transaction.CreatedAt = time.Now()
	transaction.UpdatedAt = time.Now()
	_, err := a.db.NewInsert().Model(transaction).Exec(ctx)
	return transaction, err
}

func (a *Client) GetTransaction(ctx context.Context, transactionID string) (*types.Transaction, error) {
	transaction := types.Transaction{ID: transactionID}
	err := a.db.NewSelect().Model(&transaction).WherePK().Limit(1).Scan(ctx)
	return &transaction, err
}

func (a *Client) UpdateTransaction(ctx context.Context, transaction *types.Transaction) (*types.Transaction, error) {
	transaction.UpdatedAt = time.Now()
	_, err := a.db.NewUpdate().Model(transaction).WherePK().Exec(ctx)
	return transaction, err
}

func (a *Client) UpdateTransactionStatus(
	ctx context.Context, transactionID string, status types.TransactionStatus,
) (*types.Transaction, error) {
	transaction := &types.Transaction{ID: transactionID, Status: status, UpdatedAt: time.Now()}
	_, err := a.db.
		NewUpdate().
		Model(transaction).
		Column("status", "updated_at").WherePK().Exec(ctx)
	return transaction, err
}

func (a *Client) ListTransactions(ctx context.Context, opts types.ListTransactionsFilterOpts) ([]types.Transaction, string, error) {
	var results []types.Transaction

	if opts.Limit < 1 {
		opts.Limit = 1
	}

	if opts.Limit > 500 {
		opts.Limit = 500
	}

	q := a.db.NewSelect().Model(&results).Limit(opts.Limit + 1)

	{ // filters
		if opts.CustomerID != "" {
			q.Where("customer_id = ?", opts.CustomerID)
		}

		if opts.WalletID != "" {
			q.Where("wallet_id = ?", opts.WalletID)
		}

		if len(opts.Currency) > 0 {
			q.Where("lower(currency) IN (?)", bun.In(types.ToLowercaseSlice(opts.Currency)))
		}

		if opts.IsDebit != nil {
			q.Where("is_debit = ?", opts.IsDebit)
		}

		if opts.Type != nil {
			q.Where("lower(type) = lower(?)", opts.Type)
		}

		if opts.Status != nil {
			q.Where("lower(status) = lower(?)", opts.Status)
		}

		if opts.Reversed != nil {
			if *opts.Reversed {
				q.Where("reversed_at IS NOT NULL")
			}
		}

		if !opts.StartDate.IsZero() {
			q.Where("created_at >= ?", opts.StartDate)
		}

		if !opts.EndDate.IsZero() {
			q.Where("created_at <= ?", opts.EndDate)
		}

		if strings.EqualFold(opts.Direction, types.DirectionDesc) {
			q.OrderExpr("created_at DESC")
		} else {
			q.OrderExpr("created_at ASC")
		}

	}

	if err := q.Scan(ctx); err != nil {
		return nil, "", err
	}

	var nextCursor string
	if len(results) > opts.Limit {
		nextCursor = results[opts.Limit].CreatedAt.Format(time.RFC3339)
		results = results[:opts.Limit-1]
	}

	return results, nextCursor, nil
}
