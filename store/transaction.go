package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/otyang/waas-go/types"
	"github.com/shopspring/decimal"
	"github.com/uptrace/bun"
)

// FindTransactionByID retrieves a transaction by its ID
func (r *WalletRepository) FindTransactionByID(ctx context.Context, id string) (*types.TransactionHistory, error) {
	tx := types.TransactionHistory{ID: id}

	err := r.db.NewSelect().
		Model(&tx).
		WherePK().
		Limit(1).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("transaction not found")
		}
		return nil, err
	}

	return &tx, nil
}

// UpdateTransaction updates an existing transaction
func (r *WalletRepository) UpdateTransaction(ctx context.Context, tx *types.TransactionHistory) (*types.TransactionHistory, error) {
	if tx == nil {
		return nil, errors.New("transaction cannot be nil")
	}

	// Ensure we're not updating a completed transaction
	existing, err := r.FindTransactionByID(ctx, tx.ID)
	if err != nil {
		return nil, err
	}

	if existing.Status == types.StatusCompleted {
		return nil, types.ErrTransactionCompleted
	}

	// Validate status transition
	if !existing.CanTransitionTo(tx.Status) {
		return nil, types.ErrInvalidStatusTransition
	}

	tx.UpdatedAt = time.Now().UTC()

	_, err = r.db.NewUpdate().
		Model(tx).
		WherePK().
		Exec(ctx)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

// CreateTransaction creates a new transaction record in the database
func (r *WalletRepository) CreateTransaction(ctx context.Context, txData *types.TransactionHistory) (*types.TransactionHistory, error) {
	// Validate input
	if txData == nil {
		return nil, errors.New("transaction data cannot be nil")
	}

	// Set default values if not provided
	txData.CreatedAt = time.Now().UTC()
	txData.UpdatedAt = txData.CreatedAt

	// Validate required fields
	if txData.WalletID == "" {
		return nil, errors.New("wallet ID is required")
	}
	if txData.Amount.LessThanOrEqual(decimal.Zero) {
		return nil, errors.New("transaction amount must be positive")
	}

	// Validate category
	switch txData.Category {
	case types.CategoryDeposit, types.CategoryTransfer, types.CategoryRefund,
		types.CategoryAdjustment, types.CategoryFee:
		// Valid category
	default:
		return nil, errors.New("invalid transaction category")
	}

	// Validate type
	switch txData.Type {
	case types.TypeCredit, types.TypeDebit:
		// Valid type
	default:
		return nil, errors.New("invalid transaction type")
	}

	// Validate status
	switch txData.Status {
	case types.StatusPending, types.StatusCompleted, types.StatusFailed:
		// Valid status
	default:
		return nil, errors.New("invalid transaction status")
	}

	// Insert into database
	_, err := r.db.NewInsert().
		Model(txData).
		Exec(ctx)

	return txData, err
}

// ListTransactionsParams contains parameters for listing transactions
type ListTransactionsParams struct {
	Cursor       string                    // The cursor value (usually transaction ID or created_at)
	PageSize     int                       // Number of items per page
	WalletID     string                    // Filter by wallet ID
	CurrencyCode string                    // Filter by currency code
	Category     types.TransactionCategory // Filter by transaction category
	Status       types.TransactionStatus   // Filter by transaction status
	StartTime    time.Time                 // Filter transactions after this time
	EndTime      time.Time                 // Filter transactions before this time
	SortBy       string                    // Field to sort by ("id", "created_at", "amount")
	SortOrder    string                    // Sort order ("asc" or "desc")
}

// ListTransactionsResult contains the paginated transaction results
type ListTransactionsResult struct {
	Transactions []*types.TransactionHistory `json:"transactions"`
	NextCursor   string                      `json:"next_cursor,omitempty"`
	HasNext      bool                        `json:"has_next"`
}

// ListTransactions retrieves transactions using cursor-based pagination
func (r *WalletRepository) ListTransactions(ctx context.Context, params ListTransactionsParams) (*ListTransactionsResult, error) {
	// Validate and set defaults
	if params.PageSize < 1 || params.PageSize > 100 {
		params.PageSize = 25 // Default page size with reasonable limit
	}

	// Validate sort field
	switch params.SortBy {
	case "id", "created_at", "amount", "updated_at":
		// Valid sort fields
	default:
		params.SortBy = "created_at" // Default to most recent first
	}

	// Validate sort order
	params.SortOrder = strings.ToLower(params.SortOrder)
	if params.SortOrder != "asc" && params.SortOrder != "desc" {
		params.SortOrder = "desc" // Default to descending
	}

	// Build base query
	query := r.db.NewSelect().
		Model((*types.TransactionHistory)(nil)).
		Limit(params.PageSize + 1) // Fetch one extra to check for hasNext

	// Apply filters
	if params.WalletID != "" {
		query = query.Where("wallet_id = ?", params.WalletID)
	}
	if params.CurrencyCode != "" {
		query = query.Where("currency_code = ?", params.CurrencyCode)
	}
	if params.Category != "" {
		query = query.Where("category = ?", params.Category)
	}
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}
	if !params.StartTime.IsZero() {
		query = query.Where("created_at >= ?", params.StartTime)
	}
	if !params.EndTime.IsZero() {
		query = query.Where("created_at <= ?", params.EndTime)
	}

	// Apply cursor condition
	if params.Cursor != "" {
		cursorOp := ">"
		if params.SortOrder == "desc" {
			cursorOp = "<"
		}

		// Handle different sort fields
		switch params.SortBy {
		case "created_at":
			cursorTime, err := time.Parse(time.RFC3339Nano, params.Cursor)
			if err == nil {
				query = query.Where("created_at "+cursorOp+" ?", cursorTime)
			}
		case "amount":
			cursorAmount, err := decimal.NewFromString(params.Cursor)
			if err == nil {
				query = query.Where("amount "+cursorOp+" ?", cursorAmount)
			}
		default: // Default to ID
			query = query.Where("id "+cursorOp+" ?", params.Cursor)
		}
	}

	// Apply sorting
	query = query.OrderExpr("? ?", bun.Ident(params.SortBy), bun.Safe(params.SortOrder))

	// Execute query
	var transactions []*types.TransactionHistory
	err := query.Scan(ctx, &transactions)
	if err != nil {
		return nil, fmt.Errorf("failed to list transactions: %w", err)
	}

	// Determine if there are more items
	hasNext := len(transactions) > params.PageSize
	if hasNext {
		transactions = transactions[:params.PageSize] // Remove extra item
	}

	// Get cursor for next page
	var nextCursor string
	if len(transactions) > 0 {
		lastTx := transactions[len(transactions)-1]

		switch params.SortBy {
		case "created_at":
			nextCursor = lastTx.CreatedAt.Format(time.RFC3339Nano)
		case "amount":
			nextCursor = lastTx.Amount.String()
		default:
			nextCursor = lastTx.ID
		}
	}

	return &ListTransactionsResult{
		Transactions: transactions,
		NextCursor:   nextCursor,
		HasNext:      hasNext,
	}, nil
}

// UpdateTransactionStatus updates a transaction's status with state transition validation
func (r *WalletRepository) UpdateTransactionStatus(
	ctx context.Context,
	transactionID string,
	newStatus types.TransactionStatus,
) (*types.TransactionHistory, error) {
	// Validate the new status
	switch newStatus {
	case types.StatusPending, types.StatusCompleted, types.StatusFailed:
		// Valid status
	default:
		return nil, errors.New("invalid transaction status")
	}

	// Retrieve current transaction
	currentTx, err := r.FindTransactionByID(ctx, transactionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("transaction not found")
		}
		return nil, err
	}

	// Validate status transition
	if !currentTx.CanTransitionTo(newStatus) {
		return nil, types.ErrInvalidStatusTransition
	}

	// Prepare update
	currentTx.Status = newStatus
	currentTx.UpdatedAt = time.Now().UTC()

	r.UpdateTransaction(ctx, currentTx)
	if err != nil {
		return nil, err
	}

	// Return updated transaction
	updatedTx := currentTx
	updatedTx.Status = newStatus
	updatedTx.UpdatedAt = currentTx.UpdatedAt

	return updatedTx, nil
}
