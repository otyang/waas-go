package transfer

import (
	"context"
	"time"

	"github.com/otyang/waas-go/types"
	"github.com/shopspring/decimal"
	"github.com/uptrace/bun"
)

const TransactionTypeTransfer types.TransactionCategory = "TRANSFER"

type Transfer struct {
	ID                  string                  `json:"id" bun:"id,pk"`
	CustomerID          string                  `json:"customerId" bun:",notnull"`
	SourceWalletID      string                  `json:"sourceWalletId" bun:",notnull"`
	DestinationWalletID string                  `json:"destinationWalletId" bun:",notnull"`
	Currency            string                  `json:"currency" bun:",notnull"`
	Amount              decimal.Decimal         `json:"amount" bun:"type:decimal(24,8),notnull"`
	Fee                 decimal.Decimal         `json:"fee" bun:"type:decimal(24,8),notnull"`
	Total               decimal.Decimal         `json:"total" bun:"type:decimal(24,8),notnull"`
	Status              types.TransactionStatus `json:"status" bun:",notnull"`
	InitiatorID         string                  `json:"initiatorId" bun:",notnull"`
	Narration           string                  `json:"narration" bun:",notnull"`
	CreatedAt           time.Time               `json:"createdAt" bun:",notnull"`
	UpdatedAt           time.Time               `json:"updatedAt" bun:",notnull"`
}

type ListTransferParams struct {
	CustomerID          string                  `json:"customerId"`
	SourceWalletID      string                  `json:"sourceWalletId"`
	DestinationWalletID string                  `json:"destinationWalletId"`
	Currency            string                  `json:"currency"`
	Amount              decimal.Decimal         `json:"amount"`
	Narration           string                  `json:"narration"`
	Status              types.TransactionStatus `json:"status"`
	StartDate           time.Time               `json:"createdAt"`
	EndDate             time.Time               `json:"updatedAt"`
}

func NewWithMigration(ctx context.Context, db *bun.DB) (*Client, error) {
	_, err := db.NewCreateTable().Model((*Transfer)(nil)).IfNotExists().Exec(ctx)
	if err != nil {
		return nil, err
	}

	_, err = db.NewCreateIndex().Model((*Transfer)(nil)).Index("transfer_id_index").Column("transfer_id").Exec(ctx)

	return New(db), err
}

func (t *Transfer) ToTransaction(fromWalletID, toWalletID *types.Wallet) (fromTx, toTx *types.Transaction) {

	fromWallet, err := a.FindWalletByID(ctx, opts.WalletID)
	if err != nil {
		return nil, err
	}

	toWallet, err := a.FindWalletByID(ctx, opts.WalletID)
	if err != nil {
		return nil, err
	}

	from, to, err := types.TransferWithTxn(fromWallet, toWallet, types.CreditOrDebitWalletOption{})
	if err != nil {
		return nil, err
	}

	//de.SetServiceTxnID(t.ID, false)
	// ce.SetServiceTxnID(t.ID, false)

	return de, ce
}
