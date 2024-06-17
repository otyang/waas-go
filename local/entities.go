package local

import (
	"context"
	"time"

	"github.com/otyang/waas-go/types"
	"github.com/shopspring/decimal"
	"github.com/uptrace/bun"
)

const (
	TransactionTypeLocal types.TransactionType = "local_payment"
)

type LocalPayment struct {
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
	BankName          string                  `json:"bankName" bun:",notnull"`
	BankAccountNumber string                  `json:"bankAccountNumber" bun:",notnull"`
	BankAccountName   string                  `json:"bankAccountName" bun:",notnull"`
	Description       string                  `json:"description" bun:",notnull"`
	Provider          string                  `json:"provider" bun:",notnull"`
	ProviderReference string                  `json:"providerReference"`
	CounterpartyID    *string                 `json:"counterpartyId"`
	ReversedAt        *time.Time              `json:"reversedAt"`
	CreatedAt         time.Time               `json:"createdAt" bun:",notnull"`
	UpdatedAt         time.Time               `json:"updatedAt" bun:",notnull"`
}

type ListLocalPaymentParams struct {
	WalletID          string                  `json:"walletId"`
	CustomerID        string                  `json:"customerId"`
	IsDeposit         *bool                   `json:"isDeposit"`
	Currency          string                  `json:"currency"`
	Status            types.TransactionStatus `json:"status"`
	InitiatorID       string                  `json:"initiatorId"`
	BankName          string                  `json:"bankName"`
	BankAccountNumber string                  `json:"bankAccountNumber"`
	BankAccountName   string                  `json:"bankAccountName"`
	Description       string                  `json:"description"`
	Provider          string                  `json:"provider"`
	CounterpartyID    *string                 `json:"counterpartyId"`
	Reversed          *bool                   `json:"reversedAt"`
	CreatedAt         time.Time               `json:"createdAt"`
	UpdatedAt         time.Time               `json:"updatedAt"`
}

func NewWithMigration(ctx context.Context, db *bun.DB) (*Client, error) {
	_, err := db.NewCreateTable().Model((*LocalPayment)(nil)).IfNotExists().Exec(ctx)
	return New(db), err
}

func (t *LocalPayment) ToTransaction(wallet *types.Wallet) *types.Transaction {

	return types.NewTransactionSummary(types.TxnSummaryParams{
		TransactionID:     t.ID,
		IsDebit:           !t.IsDeposit,
		Wallet:            wallet,
		Amount:            t.Amount,
		Fee:               t.Fee,
		TotalAmount:       t.Total,
		TransactionType:   TransactionTypeLocal,
		TransactionStatus: t.Status,
		Narration:         t.Description,
	})
}
