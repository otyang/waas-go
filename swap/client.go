package swap

import (
	"context"
	"time"

	"github.com/otyang/waas-go/account"
	"github.com/otyang/waas-go/types"
	"github.com/uptrace/bun"
)

type Client struct {
	db      bun.IDB
	account *account.Client
}

func New(db *bun.DB) *Client {
	return &Client{db: db}
}

func (a *Client) NewWithTx(tx bun.Tx) *Client {
	return &Client{db: tx}
}

func (a *Client) Get(ctx context.Context, swapID string) (*Swap, error) {
	swapOBJ := Swap{ID: swapID}
	err := a.db.NewSelect().Model(&swapOBJ).WherePK().Limit(1).Scan(ctx)
	return &swapOBJ, err
}

func (a *Client) Update(ctx context.Context, swapRequest *Swap) (*Swap, error) {
	swapRequest.UpdatedAt = time.Now()
	_, err := a.db.NewUpdate().Model(swapRequest).WherePK().Exec(ctx)
	return swapRequest, err
}

func (a *Client) List(ctx context.Context, opts ListSwapParams) ([]Swap, error) {
	var swaps []Swap

	q := a.db.NewSelect().Model(&swaps)

	if opts.Status != "" {
		q.Where("lower(status) = ?", opts.Status)
	}

	if opts.CustomerID != "" {
		q.Where("customer_id = ?", opts.CustomerID)
	}

	if opts.SourceCurrencyCode != "" {
		q.Where("lower(source_currency_code) = ?", opts.SourceCurrencyCode)
	}

	if opts.DestinationCurrencyCode != "" {
		q.Where("lower(destination_currency_code) = ?", opts.DestinationCurrencyCode)
	}

	if !opts.StartAmountRange.IsZero() {
		q.Where("amount => ?", opts.StartAmountRange)
	}

	if !opts.EndAmountRange.IsZero() {
		q.Where("amount <= ?", opts.EndAmountRange)
	}

	if !opts.StartDate.IsZero() {
		q.Where("created_at >= ?", opts.StartDate)
	}

	if !opts.EndDate.IsZero() {
		q.Where("created_at <= ?", opts.EndDate)
	}

	err := q.Scan(ctx)
	return swaps, err
}

func (a *Client) Create(ctx context.Context, opts *Swap) (*Swap, error) {

	sourceWallet, err := a.account.GetWalletByCurrencyCode(ctx, opts.SourceCurrencyCode, opts.CustomerID)
	if err != nil {
		return nil, err
	}

	destinationWallet, err := a.account.GetWalletByCurrencyCode(ctx, opts.DestinationCurrencyCode, opts.CustomerID)
	if err != nil {
		return nil, err
	}

	err = sourceWallet.Swap(destinationWallet, opts.FromAmount, opts.ToAmount, opts.FromFee)
	if err != nil {
		return nil, err
	}

	opts.SourceWalletID = sourceWallet.ID
	opts.DestinationWalletID = destinationWallet.ID

	fromTrsn, toTrsn := opts.ToTransaction(sourceWallet, destinationWallet)

	if err = a.account.WithTxUpdateWalletAndUpsertEvents(
		ctx,
		[]*types.Wallet{sourceWallet, destinationWallet},
		opts, fromTrsn, toTrsn,
	); err != nil {
		return nil, err
	}

	return opts, nil
}
