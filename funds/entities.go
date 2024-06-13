package funds

import (
	"context"
	"time"

	"github.com/otyang/waas-go/types"
	"github.com/shopspring/decimal"
	"github.com/uptrace/bun"
)

type Fiat struct {
	ID                string                  `json:"id" bun:"id,pk"`
	WalletID          string                  `json:"walletId" bun:",notnull"`
	CustomerID        string                  `json:"customerId" bun:",notnull"`
	IsDeposit         bool                    `json:"isDeposit" bun:",notnull"`
	Currency          string                  `json:"currency" bun:",notnull"`
	Amount            decimal.Decimal         `json:"amount" bun:",notnull"`
	Fee               decimal.Decimal         `json:"fee" bun:",notnull"`
	Total             decimal.Decimal         `json:"total" bun:"type:decimal(24,8),notnull"`
	Status            types.TransactionStatus `json:"status" bun:",notnull"`
	InitiatorID       string                  `json:"initiatorId" bun:",notnull"`
	CreatedAt         time.Time               `json:"createdAt" bun:",notnull"`
	UpdatedAt         time.Time               `json:"updatedAt" bun:",notnull"`
	BankName          *string                 `json:"bankName"`
	BankAccountNo     string                  `json:"bankAccountNumber" bun:",notnull"`
	BankAccountName   string                  `json:"bankAccountName"`
	ProviderReference string                  `json:"providerReference"`
}

func NewWithMigration(ctx context.Context, db *bun.DB) (*Client, error) {
	_, err := db.NewCreateTable().Model((*Fiat)(nil)).IfNotExists().Exec(ctx)
	return New(db), err
}

func (t *Fiat) ToTransaction(wallet *types.Wallet, txStatus types.TransactionStatus) *types.Transaction {
	tx := types.NewTransaction(t.IsDeposit, wallet, t.Amount, t.Fee, t.Amount.Add(t.Fee), types.TransactionTypeDeposit, txStatus)
	return tx
}
