package funds

import (
	"context"

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

func (a *Client) Fiat(ctx context.Context, opts *Fiat) (*Fiat, error) {
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

	transaction := opts.ToTransaction(wallet, types.TransactionStatusSuccess)

	if err = a.account.WithTxUpdateWalletAndUpsertEvents(ctx, []*types.Wallet{wallet}, opts, transaction); err != nil {
		return nil, err
	}

	return opts, nil
}
