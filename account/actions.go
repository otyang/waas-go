package account

import (
	"context"

	"github.com/otyang/waas-go/types"
	"github.com/uptrace/bun"
)

func (a *Client) WithTxBulkUpdateWalletAndInsertTransaction(
	ctx context.Context,
	wallets []*types.Wallet,
	transactions []*types.Transaction,
	otherEntries ...any,
) error {
	return a.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		for _, wallet := range wallets {
			if wallet == nil {
				continue
			}
			_, err := a.NewWithTx(tx).UpdateWallet(ctx, wallet)
			if err != nil {
				return err
			}
		}

		if len(transactions) > 0 {
			for _, transaction := range transactions {
				if transaction == nil {
					continue
				}
				_, err := a.NewWithTx(tx).CreateTransaction(ctx, transaction)
				if err != nil {
					return err
				}
			}
		}

		// for other entries. transaction specific structs
		if len(otherEntries) > 0 {
			for _, otherEntry := range otherEntries {
				if otherEntry == nil {
					continue
				}
				_, err := tx.NewInsert().Model(otherEntry).Exec(ctx)
				if err != nil {
					return err
				}
			}
		}

		return nil
	})
}
