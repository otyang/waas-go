package account

import (
	"context"
	"sort"
	"strings"
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

func (a *Account) ListTransaction(ctx context.Context, params waas.ListTransactionsFilterParams) ([]waas.Transaction, error) {
	var transactions []waas.Transaction
	var beforeOnly bool

	if params.Limit < 1 {
		params.Limit = 1
	}

	if params.Limit > 500 {
		params.Limit = 500
	}

	var sqlOrder string = " desc"
	if params.SortOrderAscending {
		sqlOrder = " asc"
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
			q.Where("lower(currency) IN (?)", params.Currency)
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
			q.Where("reversed IS NOT NULL")
		}

		if params.Before.IsZero() && !params.After.IsZero() {
			q.Where("created_at >= ?", params.After).OrderExpr("created_at " + sqlOrder)
		}

		if !params.Before.IsZero() && params.After.IsZero() {
			beforeOnly = true
			q.Where("created_at <= ?", params.Before).OrderExpr("created_at " + sqlOrder)
		}

		// default case: transactions before 2days
		if params.Before.IsZero() && params.After.IsZero() {
			beforeOnly = true
			q.Where("created_at <= ?", time.Now().UTC().Add(24*2*time.Hour)).OrderExpr("created_at " + sqlOrder)
		}
	}

	err := q.Scan(ctx)
	if err != nil {
		return nil, err
	}

	if beforeOnly && strings.Contains(sqlOrder, "desc") {
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
