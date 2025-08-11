package store

import (
	"github.com/uptrace/bun"
)

type WalletRepository struct {
	db bun.IDB
}

func NewWalletRepository(db *bun.DB) *WalletRepository {
	return &WalletRepository{db: db}
}

func (a *WalletRepository) NewWithTx(tx bun.Tx) *WalletRepository {
	return &WalletRepository{db: tx}
}
