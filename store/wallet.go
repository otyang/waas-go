package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/otyang/waas-go/types"
	"github.com/uptrace/bun"
)

var (
	// ErrWalletAlreadyExists    = errors.New("wallet already exists")
	ErrWalletNotFound         = errors.New("wallet not found")
	ErrConcurrentModification = errors.New("wallet was modified concurrently")
	ErrCustomerIDRequired     = errors.New("customer ID is required")
	ErrInvalidCurrencyCode    = errors.New("invalid currency code")
)

// CreateWallet creates a new wallet or returns existing one
func (c *WalletRepository) CreateWallet(ctx context.Context, wallet *types.Wallet) (*types.Wallet, error) {
	if wallet == nil {
		return nil, errors.New("wallet cannot be nil")
	}

	// Normalize currency code
	wallet.CurrencyCode = strings.ToUpper(strings.TrimSpace(wallet.CurrencyCode))

	// Check for existing wallet
	existing, err := c.FindWalletByCurrency(ctx, wallet.CustomerID, wallet.CurrencyCode)

	if err == nil {
		return existing, nil
	}

	if err == sql.ErrNoRows {
		// Initialize wallet fields if empty
		if wallet.ID == "" {
			wallet.ID = types.GenerateID("wt_", 12)
		}
		if wallet.VersionId == "" {
			wallet.VersionId = types.GenerateID("ver_", 8)
		}
		if wallet.CreatedAt.IsZero() {
			wallet.CreatedAt = time.Now().UTC()
		}
		wallet.UpdatedAt = time.Now().UTC()
	}

	// Insert new wallet
	_, err = c.db.NewInsert().
		Model(wallet).
		Ignore().
		Exec(ctx)

	return wallet, err
}

// CreateSimplified creates a new wallet with minimal parameters
func (c *WalletRepository) CreateSimplified(ctx context.Context, customerID, currencyCode string) (*types.Wallet, error) {
	if strings.TrimSpace(customerID) == "" {
		return nil, ErrCustomerIDRequired
	}

	currencyCode = strings.TrimSpace(strings.ToUpper(currencyCode))
	if currencyCode == "" {
		return nil, ErrInvalidCurrencyCode
	}

	wallet, err := types.NewWallet(customerID, currencyCode)
	if err != nil {
		return nil, err
	}

	return c.CreateWallet(ctx, wallet)
}

// UpdateWallet updates an existing wallet with optimistic concurrency control
func (c *WalletRepository) UpdateWallet(ctx context.Context, wallet *types.Wallet) (*types.Wallet, error) {
	if wallet == nil {
		return nil, errors.New("wallet cannot be nil")
	}

	oldVersion := wallet.VersionId
	wallet.VersionId = types.GenerateID("ver_", 8)
	wallet.UpdatedAt = time.Now().UTC()

	res, err := c.db.NewUpdate().
		Model(wallet).
		WherePK().
		Where("version_id = ?", oldVersion).
		Exec(ctx)
	if err != nil {
		return nil, err
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return nil, ErrConcurrentModification
	}

	return wallet, nil
}

// FindWalletByID retrieves a wallet by its ID
func (c *WalletRepository) FindWalletByID(ctx context.Context, walletID string) (*types.Wallet, error) {
	wallet := &types.Wallet{ID: walletID}

	err := c.db.NewSelect().
		Model(wallet).
		WherePK().
		Limit(1).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrWalletNotFound
		}
		return nil, err
	}

	return wallet, nil
}

// FindWalletByCurrency retrieves a wallet by customer ID and currency code
func (c *WalletRepository) FindWalletByCurrency(ctx context.Context, customerID, currencyCode string) (*types.Wallet, error) {
	wallet := new(types.Wallet)
	err := c.db.NewSelect().
		Model(wallet).
		Where("customer_id = ? AND currency_code = ?", customerID, strings.ToUpper(currencyCode)).
		Limit(1).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrWalletNotFound
		}
		return nil, err
	}

	return wallet, nil
}

// ListWalletsParams contains parameters for listing wallets
type ListWalletsParams struct {
	CustomerID   string   // Filter by customer ID
	CurrencyCode []string // Filter by one or more currency codes
	IsFrozen     *bool    // Filter by frozen status (nil for no filter)
	IsClosed     *bool    // Filter by closed status (nil for no filter)
	Page         int      // Page number (1-based)
	PageSize     int      // Number of items per page
	SortBy       string   // Field to sort by (e.g., "created_at", "balance")
	SortOrder    string   // Sort order ("asc" or "desc")
}

// ListWalletsResult contains the paginated wallet listing results
type ListWalletsResult struct {
	Wallets     []*types.Wallet `json:"wallets"`
	TotalCount  int             `json:"total_count"`
	CurrentPage int             `json:"current_page"`
	TotalPages  int             `json:"total_pages"`
	HasNext     bool            `json:"has_next"`
	HasPrevious bool            `json:"has_previous"`
}

// ListWalletsParams contains parameters for cursor-based wallet listing
type ListWalletsParamsCursor struct {
	Cursor       string   // The cursor pointing to the last item of the previous page
	PageSize     int      // Number of items per page
	CustomerID   string   // Filter by customer ID
	CurrencyCode []string // Filter by one or more currency codes
	IsFrozen     *bool    // Filter by frozen status
	IsClosed     *bool    // Filter by closed status
	SortBy       string   // Field to sort by (must be unique, like "id" or "created_at")
	SortOrder    string   // Sort order ("asc" or "desc")
}

// ListWalletsResult contains the cursor-paginated wallet listing results
type ListWalletsResultCursor struct {
	Wallets    []*types.Wallet `json:"wallets"`
	NextCursor string          `json:"next_cursor,omitempty"`
	HasNext    bool            `json:"has_next"`
	TotalCount int             `json:"total_count,omitempty"` // Optional, expensive for large datasets
}

// ListWallets retrieves wallets using cursor-based pagination
func (c *WalletRepository) ListWalletsCursor(ctx context.Context, params ListWalletsParamsCursor) (*ListWalletsResultCursor, error) {
	// Validate parameters
	if params.PageSize < 1 || params.PageSize > 100 {
		params.PageSize = 20 // Default page size with reasonable upper limit
	}

	// Ensure we have a valid sort field - must be unique for cursor pagination
	switch params.SortBy {
	case "id", "created_at", "updated_at":
		// These are good unique fields for cursor pagination
	default:
		params.SortBy = "created_at" // Fallback to default
	}

	// Validate sort order
	params.SortOrder = strings.ToLower(params.SortOrder)
	if params.SortOrder != "asc" && params.SortOrder != "desc" {
		params.SortOrder = "desc" // Default sort order
	}

	// Build base query
	query := c.db.NewSelect().
		Model((*types.Wallet)(nil)).
		Limit(params.PageSize + 1) // Fetch one extra to check for hasNext

	// Apply filters
	if params.CustomerID != "" {
		query = query.Where("customer_id = ?", params.CustomerID)
	}
	if len(params.CurrencyCode) > 0 {
		query = query.Where("currency_code IN (?)", bun.In(params.CurrencyCode))
	}
	if params.IsFrozen != nil {
		query = query.Where("is_frozen = ?", *params.IsFrozen)
	}
	if params.IsClosed != nil {
		query = query.Where("is_closed = ?", *params.IsClosed)
	}

	// Apply cursor condition
	if params.Cursor != "" {
		if params.SortOrder == "asc" {
			query = query.Where("? > ?", bun.Ident(params.SortBy), params.Cursor)
		} else {
			query = query.Where("? < ?", bun.Ident(params.SortBy), params.Cursor)
		}
	}

	// Apply sorting
	query = query.OrderExpr("? ?", bun.Ident(params.SortBy), bun.Safe(params.SortOrder))

	// Execute query
	var wallets []*types.Wallet
	err := query.Scan(ctx, &wallets)
	if err != nil {
		return nil, fmt.Errorf("failed to list wallets: %w", err)
	}

	// Determine if there are more items
	hasNext := false
	var nextCursor string

	if len(wallets) > params.PageSize {
		hasNext = true
		wallets = wallets[:params.PageSize] // Remove the extra item
	}

	// Get the cursor for the next page
	if len(wallets) > 0 {
		lastItem := wallets[len(wallets)-1]

		// Use reflection to get the sort field dynamically
		val := reflect.ValueOf(lastItem).Elem()
		field := val.FieldByNameFunc(func(name string) bool {
			return strings.EqualFold(name, params.SortBy)
		})

		if field.IsValid() {
			nextCursor = fmt.Sprintf("%v", field.Interface())
		}
	}

	// Note: TotalCount is omitted by default as it's expensive for cursor pagination
	// You could add it optionally with a separate count query if needed

	return &ListWalletsResultCursor{
		Wallets:    wallets,
		NextCursor: nextCursor,
		HasNext:    hasNext,
	}, nil
}
