package account

import (
	"context"
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
	oldIdempotencyID := transaction.IdempotencyId  // extract oldVersionID. for concurrency locks
	transaction.IdempotencyId = waas.GenerateID(7) // newVId
	transaction.UpdatedAt = time.Now()

	_, err := a.db.NewUpdate().
		Model(transaction).WherePK().
		Where("idempotency_id = ?", oldIdempotencyID).
		Exec(ctx)

	return transaction, err
}

func (a *Account) ListTransaction(ctx context.Context, limit int, params waas.ListTransactionsFilterParams) ([]waas.Transaction, error) {
	var transactions []waas.Transaction

	if limit < 1 {
		limit = 1
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
			q.Where("currency = ?", params.Currency)
		}

		if params.IsDebit != nil {
			q.Where("is_debit = ?", params.IsDebit)
		}

		if params.Type != nil {
			q.Where("type = ?", params.Type)
		}

		if params.Status != nil {
			q.Where("status = ?", params.Status)
		}

		if params.Reversed != nil {
			q.Where("reversed = ?", params.Reversed)
		}

		if !params.After.IsZero() {
			q.Where("created_at >= ?", params.After).OrderExpr("created_at ASC")
		}

		if !params.Before.IsZero() {
			q.Where("created_at <= ?", params.Before).OrderExpr("created_at DESC")
		}

		// default case
		if params.After.IsZero() && params.Before.IsZero() {
			q.Where("created_at <= ?", time.Now().Add(24*2*time.Hour)).OrderExpr("created_at DESC")
		}
	}

	err := q.Scan(ctx)
	if err != nil {
		return nil, err
	}

	return transactions, nil
	// return transactions, newCursor(transactions, limit), nil
}

func newCursor(tns []waas.Transaction, limit int) {
	if len(tns) == 0 {
		return // noCursor
	}

	if len(tns) <= limit {
		return // o, limit-1, hasMore=false
	}

	return // 0, limit , hasMore=true
}
