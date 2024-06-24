package swap

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/otyang/waas-go/types"
	"github.com/shopspring/decimal"
	"github.com/uptrace/bun"
)

const (
	TransactionTypeSwap types.TransactionType = "SWAP"
)

type Swap struct {
	ID                      string                  `json:"id" bun:"id,pk"`
	CustomerID              string                  `json:"customerId" bun:",notnull"`
	SourceWalletID          string                  `json:"sourceWalletId" bun:",notnull"`
	SourceCurrencyCode      string                  `json:"sourceCurrency" bun:",notnull"`
	FromAmount              decimal.Decimal         `json:"fromAmount" bun:",notnull"`
	FromFee                 decimal.Decimal         `json:"toFee" bun:",notnull"`
	DestinationCurrencyCode string                  `json:"destinationCurrency" bun:",notnull"`
	DestinationWalletID     string                  `json:"destinationWalletId" bun:",notnull"`
	ToAmount                decimal.Decimal         `json:"toAmount" bun:",notnull"`
	Status                  types.TransactionStatus `json:"status" bun:",notnull"`
	CreatedAt               time.Time               `json:"createdAt" bun:",notnull"`
	UpdatedAt               time.Time               `json:"updatedAt" bun:",notnull"`
}

type ListSwapParams struct {
	CustomerID              string                  `json:"customerId"`
	SourceCurrencyCode      string                  `json:"sourceCurrency"`
	DestinationCurrencyCode string                  `json:"destinationCurrency"`
	StartAmountRange        decimal.Decimal         `json:"startAmountRange"`
	EndAmountRange          decimal.Decimal         `json:"endAmountRange"`
	Status                  types.TransactionStatus `json:"status"`
	StartDate               time.Time               `json:"createdAt"`
	EndDate                 time.Time               `json:"updatedAt"`
}

func NewWithMigration(ctx context.Context, db *bun.DB) (*Client, error) {
	_, err := db.NewCreateTable().Model((*Swap)(nil)).IfNotExists().Exec(ctx)
	return New(db), err
}

func (t *Swap) ToTransaction(fromWallet, toWallet *types.Wallet) (fromTx, toTx *types.Transaction) {
	de := types.NewTransactionSummary(types.TxnSummaryParams{
		IsDebit:           true,
		Wallet:            fromWallet,
		Amount:            t.FromAmount,
		Fee:               t.FromFee,
		TotalAmount:       t.FromAmount.Add(t.FromFee),
		TransactionType:   TransactionTypeSwap,
		TransactionStatus: types.TransactionStatusSuccess,
		Narration:         generateSwapNarration(t.SourceCurrencyCode, t.DestinationCurrencyCode),
		CreatedAt:         t.CreatedAt,
		UpdatedAt:         t.UpdatedAt,
	})

	ce := types.NewTransactionSummary(types.TxnSummaryParams{
		IsDebit:           false,
		Wallet:            toWallet,
		Amount:            t.ToAmount,
		Fee:               decimal.Zero,
		TotalAmount:       t.ToAmount,
		TransactionType:   TransactionTypeSwap,
		TransactionStatus: types.TransactionStatusSuccess,
		Narration:         generateSwapNarration(t.SourceCurrencyCode, t.DestinationCurrencyCode),
		CreatedAt:         t.CreatedAt,
		UpdatedAt:         t.UpdatedAt,
	})

	de.SetServiceTxnID(t.ID, false)
	ce.SetServiceTxnID(t.ID, false)

	return de, ce
}

func generateSwapNarration(fromCurrency, toCurrency string) string {
	return fmt.Sprintf("%s - %s", strings.ToLower(fromCurrency), strings.ToLower(toCurrency))
}
