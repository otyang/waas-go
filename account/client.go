package account

import (
	"context"

	"github.com/otyang/waas-go/types"
	"github.com/uptrace/bun"
)

// var _ types.IAccountFeature = (*Client)(nil)
type Client struct {
	db bun.IDB
}

func New(db *bun.DB) *Client {
	return &Client{db: db}
}

func (a *Client) NewWithTx(tx bun.Tx) *Client {
	return &Client{db: tx}
}

func NewWithMigration(db *bun.DB) (*Client, error) {
	ctx := context.Background()

	_, err := db.NewCreateTable().Model((*types.Transaction)(nil)).IfNotExists().Exec(ctx)
	if err != nil {
		return nil, err
	}

	_, err = db.NewCreateTable().Model((*types.Wallet)(nil)).IfNotExists().Exec(ctx)
	if err != nil {
		return nil, err
	}
	_, err = db.NewCreateTable().Model((*types.Currency)(nil)).IfNotExists().Exec(ctx)
	return New(db), err
}
