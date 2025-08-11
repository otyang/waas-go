package store

// import (
// 	"context"
// 	"errors"
// 	"fmt"

// 	"github.com/otyang/waas-go/types"
// )

// // UpdateWalletsAndTransactionsInsertEvents updates wallets, creates transactions, and inserts events
// // in separate auto-committed operations. This is suitable when atomicity across all operations
// // is not required.
// func (r *WalletRepository) UpdateWalletsAndTransactionsInsertEvents(
// 	ctx context.Context,
// 	wallets []*types.Wallet,
// 	transactions []*types.TransactionHistory,
// 	events []any,
// ) error {
// 	// Early return if no operations requested
// 	if len(wallets) == 0 && len(transactions) == 0 && len(events) == 0 {
// 		return errors.New("no operations requested: must provide wallets, transactions, or events")
// 	}

// 	// Process wallets
// 	for i, wallet := range wallets {
// 		if wallet == nil {
// 			continue
// 		}
// 		if _, err := r.UpdateWallet(ctx, wallet); err != nil {
// 			return fmt.Errorf("wallet update failed [index:%d, id:%s]: %w", i, wallet.ID, err)
// 		}
// 	}

// 	// Process transactions
// 	for i, tx := range transactions {
// 		if tx == nil {
// 			continue
// 		}
// 		if _, err := r.CreateTransaction(ctx, tx); err != nil {
// 			return fmt.Errorf("transaction creation failed [index:%d]: %w", i, err)
// 		}
// 	}

// 	// Process events
// 	for i, event := range events {
// 		if event == nil {
// 			continue
// 		}
// 		if _, err := r.db.NewInsert().Model(event).Exec(ctx); err != nil {
// 			return fmt.Errorf("event insertion failed [index:%d]: %w", i, err)
// 		}
// 	}

// 	return nil
// }
