package account

import (
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
	return &Account{
		db: tx,
	}
}
