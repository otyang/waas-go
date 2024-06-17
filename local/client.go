package local

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

func (a *Client) Create(ctx context.Context, opts *LocalPayment) (*LocalPayment, error) {
	wallet, err := a.account.GetWalletByID(ctx, opts.WalletID)
	if err != nil {
		return nil, err
	}

	if opts.IsDeposit {
		err = wallet.CreditBalance(opts.Amount, opts.Fee)
	} else {
		err = wallet.DebitBalance(opts.Amount, opts.Fee)
	}
	if err != nil {
		return nil, err
	}

	transaction := opts.ToTransaction(wallet)

	if err = a.account.WithTxUpdateWalletAndUpsertEvents(ctx, []*types.Wallet{wallet}, opts, transaction); err != nil {
		return nil, err
	}

	return opts, nil
}

func (a *Client) Find(ctx context.Context, txID string) (*LocalPayment, error) {
	model := LocalPayment{ID: txID}
	err := a.db.NewSelect().Model(&model).WherePK().Limit(1).Scan(ctx)
	return &model, err
}

func (a *Client) Update(ctx context.Context, m *LocalPayment) (*LocalPayment, error) {
	m.UpdatedAt = time.Now()
	_, err := a.db.NewUpdate().Model(m).WherePK().Exec(ctx)
	return m, err
}

func (a *Client) List(ctx context.Context, opts ListLocalPaymentParams) ([]LocalPayment, error) {
	var model []LocalPayment

	q := a.db.NewSelect().Model(&model)

	if opts.WalletID != "" {
		q.Where("wallet_id = ?", opts.WalletID)
	}

	if opts.CustomerID != "" {
		q.Where("customer_id = ?", opts.CustomerID)
	}

	if opts.IsDeposit != nil {
		q.Where("is_deposit = ?", opts.IsDeposit)
	}

	if opts.Currency != "" {
		q.Where("lower(currency) = ?", opts.Currency)
	}

	if opts.Status != "" {
		q.Where("lower(status) = ?", opts.Status)
	}

	if opts.InitiatorID != "" {
		q.Where("initiator_id = ?", opts.InitiatorID)
	}

	if opts.BankName != "" {
		q.Where("bank_name = ?", opts.BankName)
	}

	if opts.BankAccountName != "" {
		q.Where("bank_account_name = ?", opts.BankAccountName)
	}

	if opts.BankAccountNumber != "" {
		q.Where("bank_account_number = ?", opts.BankAccountNumber)
	}

	if opts.Description != "" {
		q.Where("description ILIKE ?", "%"+opts.Description+"%")
	}

	if opts.Provider != "" {
		q.Where("lower(provider) = ?", opts.Provider)
	}

	if opts.Reversed != nil {
		q.Where("reversed_at IS NOT NULL")
	}

	if !opts.CreatedAt.IsZero() {
		q.Where("created_at >= ?", opts.CreatedAt)
	}

	if !opts.UpdatedAt.IsZero() {
		q.Where("created_at <= ?", opts.UpdatedAt)
	}

	err := q.Scan(ctx)
	return model, err
}
