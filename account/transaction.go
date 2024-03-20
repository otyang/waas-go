package account

import (
	"context"
	"sort"
	"time"

	"github.com/otyang/waas-go"
)

func (a *Account) CreateTransaction(ctx context.Context, transaction *waas.Transaction) (*waas.Transaction, error) {
	transaction.CreatedAt = time.Now()
	transaction.UpdatedAt = time.Now()
	_, err := a.db.NewInsert().Model(transaction).Exec(ctx)
	return transaction, err
}

func (a *Account) GetTransaction(ctx context.Context, transactionID string) (*waas.Transaction, error) {
	transaction := waas.Transaction{ID: transactionID}
	err := a.db.NewSelect().Model(&transaction).WherePK().Limit(1).Scan(ctx)
	return &transaction, err
}

func (a *Account) UpdateTransaction(ctx context.Context, transaction *waas.Transaction) (*waas.Transaction, error) {
	transaction.UpdatedAt = time.Now()
	_, err := a.db.NewUpdate().Model(transaction).WherePK().Exec(ctx)
	return transaction, err
}

func (a *Account) UpdateTransactionStatus(ctx context.Context, transactionID string, status waas.TransactionStatus) (*waas.Transaction, error) {
	transaction := &waas.Transaction{ID: transactionID, Status: status, UpdatedAt: time.Now()}
	_, err := a.db.NewUpdate().Model(transaction).Column("status", "updated_at").WherePK().Exec(ctx)
	return transaction, err
}

func (a *Account) ListTransaction(ctx context.Context, limit int, params waas.ListTransactionsFilterParams) ([]waas.Transaction, error) {
	var transactions []waas.Transaction
	var sortSlice bool

	if limit < 1 {
		limit = 1
	}

	if limit > 500 {
		limit = 500
	}

	q := a.db.NewSelect().Model(&transactions).Limit(limit + 1)

	{ // filters
		if params.CustomerID != nil {
			q.Where("customer_id = ?", params.CustomerID)
		}

		if params.WalletID != nil {
			q.Where("wallet_id = ?", params.WalletID)
		}

		if params.Currency != nil {
			q.Where("lower(currency) = lower(?)", params.Currency)
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

		if params.Reversed != nil {
			q.Where("reversed = ?", params.Reversed)
		}

		if !params.After.IsZero() && params.Before.IsZero() {
			q.Where("created_at >= ?", params.After).OrderExpr("created_at ASC")
		}

		if params.After.IsZero() && !params.Before.IsZero() {
			sortSlice = true
			q.Where("created_at <= ?", params.Before).OrderExpr("created_at DESC")
		}

		// default case
		if params.After.IsZero() && params.Before.IsZero() {
			sortSlice = true
			q.Where("created_at <= ?", time.Now().UTC().Add(24*2*time.Hour)).OrderExpr("created_at DESC")
		}
	}

	err := q.Scan(ctx)
	if err != nil {
		return nil, err
	}

	// sort slice
	if sortSlice {
		sort.Slice(transactions, func(i, j int) bool {
			return transactions[i].ID < transactions[j].ID
		})
	}

	return transactions, nil
	// return transactions, newCursor(transactions, limit), nil
}

type Cursor struct {
	Start   string
	End     string
	HasMore bool
}

func newCursor(records []waas.Transaction, limit int) ([]waas.Transaction, Cursor) {
	if len(records) == 0 {
		return records, Cursor{}
	}

	if len(records) > limit {
		return records[:limit-1], Cursor{
			Start:   records[0].CreatedAt.GoString(),
			End:     records[limit].ID,
			HasMore: true,
		}
	}

	return records[:len(records)-1], Cursor{
		Start:   records[0].CreatedAt.GoString(),
		End:     records[len(records)-1].ID,
		HasMore: false,
	}
}
