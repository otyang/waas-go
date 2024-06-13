package swap

import (
	"context"
	"time"

	"github.com/otyang/waas-go/types"
	"github.com/shopspring/decimal"
	"github.com/uptrace/bun"
)

type Swap struct {
	ID                  string                  `json:"id" bun:"id,pk"`
	CustomerID          string                  `json:"customerId" bun:",notnull"`
	SourceWalletID      string                  `json:"sourceWalletId" bun:",notnull"`
	DestinationWalletID string                  `json:"destinationWalletId" bun:",notnull"`
	FromCurrencyCode    string                  `json:"fromCurrency" bun:",notnull"`
	ToCurrencyCode      string                  `json:"toCurrency" bun:",notnull"`
	FromAmount          decimal.Decimal         `json:"fromAmount" bun:",notnull"`
	FromFee             decimal.Decimal         `json:"toFee" bun:",notnull"`
	ToAmount            decimal.Decimal         `json:"toAmount" bun:",notnull"`
	Status              types.TransactionStatus `json:"status" bun:",notnull"`
	CreatedAt           time.Time               `json:"createdAt" bun:",notnull"`
	UpdatedAt           time.Time               `json:"updatedAt" bun:",notnull"`
}

type ListSwapParams struct {
	ID               string                  `json:"id"`
	CustomerID       string                  `json:"customerId"`
	FromCurrencyCode string                  `json:"fromCurrency"`
	ToCurrencyCode   string                  `json:"toCurrency"`
	StartAmountRange decimal.Decimal         `json:"startAmountRange"`
	EndAmountRange   decimal.Decimal         `json:"endAmountRange"`
	Status           types.TransactionStatus `json:"status"`
	StartDate        time.Time               `json:"createdAt"`
	EndDate          time.Time               `json:"updatedAt"`
}

func NewWithMigration(ctx context.Context, db *bun.DB) (*Client, error) {
	_, err := db.NewCreateTable().Model((*Swap)(nil)).IfNotExists().Exec(ctx)
	return New(db), err
}

func (t *Swap) ToTransaction(fromWallet, toWallet *types.Wallet, debitAmount, creditAmount, fee decimal.Decimal,
) (fromTx, toTx *types.Transaction) {
	// Since swaps are internal and always successful, set the status to success.
	de := types.NewTransaction(
		false, fromWallet, debitAmount, fee, debitAmount.Add(fee),
		types.TransactionTypeSwap, types.TransactionStatusSuccess,
	)
	ce := types.NewTransaction(
		true, toWallet, creditAmount, decimal.Zero, creditAmount.Add(decimal.Zero),
		types.TransactionTypeSwap, types.TransactionStatusSuccess,
	)

	de.SetCounterpartyID(ce.ID)
	ce.SetCounterpartyID(de.ID)

	return de, ce
}
