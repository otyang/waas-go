package account

import (
	"context"

	"github.com/otyang/waas-go/types"
	"github.com/uptrace/bun"
)

var _ types.IAccountFeature = (*Account)(nil)

type Account struct {
	db bun.IDB
}

func New(db *bun.DB) *Account {
	return &Account{db: db}
}

func (a *Account) NewWithTx(tx bun.Tx) *Account {
	return &Account{db: tx}
}

func NewWithMigration(db *bun.DB) (*Account, error) {
	ctx := context.Background()

	_, err := db.NewCreateTable().Model((*types.Transaction)(nil)).IfNotExists().Exec(ctx)
	if err != nil {
		return nil, err
	}

	_, err = db.NewCreateTable().Model((*types.Wallet)(nil)).IfNotExists().Exec(ctx)
	if err != nil {
		return nil, err
	}

	// db.NewCreateIndex().
	// 	Model((*Book)(nil)).
	// 	Index("category_id_idx").
	// 	Column("category_id").
	// 	Exec(ctx)

	_, err = db.NewCreateTable().Model((*types.Currency)(nil)).IfNotExists().Exec(ctx)
	return New(db), err
}
