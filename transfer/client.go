package transfer

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

func (a *Client) Find(ctx context.Context, txID string) (*Transfer, error) {
	transferObj := Transfer{ID: txID}
	err := a.db.NewSelect().Model(&transferObj).WherePK().Limit(1).Scan(ctx)
	return &transferObj, err
}

func (a *Client) Update(ctx context.Context, trf *Transfer) (*Transfer, error) {
	trf.UpdatedAt = time.Now()
	_, err := a.db.NewUpdate().Model(trf).WherePK().Exec(ctx)
	return trf, err
}

func (a *Client) List(ctx context.Context, opts ListTransferParams) ([]Transfer, error) {
	var trsfs []Transfer

	q := a.db.NewSelect().Model(&trsfs)

	if opts.CustomerID != "" {
		q.Where("customer_id = ?", opts.CustomerID)
	}

	if opts.SourceWalletID != "" {
		q.Where("source_wallet_id = ?", opts.SourceWalletID)
	}

	if opts.DestinationWalletID != "" {
		q.Where("destination_wallet_id = ?", opts.DestinationWalletID)
	}

	if opts.Currency != "" {
		q.Where("lower(currency) = ?", opts.Currency)
	}

	if !opts.Amount.IsZero() {
		q.Where("amount = ?", opts.Amount)
	}

	if opts.Narration != "" {
		q.Where("lower(narration) LIKE ?", "%"+opts.Narration+"%")
	}

	if opts.Status != "" {
		q.Where("lower(status) = ?", opts.Status)
	}

	if !opts.StartDate.IsZero() {
		q.Where("created_at >= ?", opts.StartDate)
	}

	if !opts.EndDate.IsZero() {
		q.Where("created_at <= ?", opts.EndDate)
	}

	err := q.Scan(ctx)
	return trsfs, err
}

func (a *Client) Create(ctx context.Context, opts *Transfer) (*Transfer, error) {
	fromWallet, err := a.account.FindWalletByID(ctx, opts.SourceWalletID)
	if err != nil {
		return nil, err
	}

	toWallet, err := a.account.FindWalletByID(ctx, opts.DestinationWalletID)
	if err != nil {
		return nil, err
	}

	err = fromWallet.TransferTo(toWallet, opts.Amount, opts.Fee)
	if err != nil {
		return nil, err
	}

	fromTrsn, toTrsn := opts.ToTransaction(fromWallet, toWallet)

	if err = a.account.WithTxUpdateWalletAndUpsertEvents(
		ctx,
		[]*types.Wallet{fromWallet, toWallet},
		opts, fromTrsn, toTrsn,
	); err != nil {
		return nil, err
	}

	return opts, nil
}
