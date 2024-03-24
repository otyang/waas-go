package account

import (
	"context"

	"github.com/otyang/waas-go"
	"github.com/uptrace/bun"
)

var _ waas.IAccountFeature = (*Account)(nil)

type Account struct {
	db bun.IDB
}

func New(db *bun.DB) *Account {
	return &Account{db: db}
}

func (a *Account) NewWithTx(tx bun.Tx) *Account {
	return &Account{db: tx}
}

func NewWithMigration(db *bun.DB) error {
	ctx := context.Background()

	_, err := db.NewCreateTable().Model((*waas.Transaction)(nil)).IfNotExists().Exec(ctx)
	if err != nil {
		return err
	}

	_, err = db.NewCreateTable().Model((*waas.Wallet)(nil)).IfNotExists().Exec(ctx)
	if err != nil {
		return err
	}

	_, err = db.NewCreateTable().Model((*waas.Currency)(nil)).IfNotExists().Exec(ctx)
	return err
}
