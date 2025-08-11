package store

import (
	"context"
	"fmt"
	"strings"

	"github.com/otyang/waas-go/types"
	"github.com/uptrace/bun"
)

// ProcessLien handles both placing and releasing liens in a single atomic operation
func (r *WalletRepository) ProcessLien(
	ctx context.Context,
	walletID string,
	request types.LienOrUnlienRequest,
	operationType string, // "lien" or "unlien"
) (*types.LienRecord, *types.Wallet, error) {
	var (
		lienRecord *types.LienRecord
		wallet     *types.Wallet
		err        error
	)

	err = r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		// Get repository with transaction
		repo := r.NewWithTx(tx)

		// 1. Retrieve wallet with lock
		wallet, err = repo.FindWalletByID(ctx, walletID)
		if err != nil {
			return fmt.Errorf("failed to get wallet: %w", err)
		}

		// 2. Perform the lien operation
		switch strings.ToLower(operationType) {
		case "lien":
			lienRecord, err = wallet.AddLien(request)
		case "unlien":
			lienRecord, err = wallet.ReleaseLien(request)
		default:
			return fmt.Errorf("invalid operation type: %s", operationType)
		}

		if err != nil {
			return fmt.Errorf("%s operation failed: %w", operationType, err)
		}

		// 3. Update wallet
		if _, err := repo.UpdateWallet(ctx, wallet); err != nil {
			return fmt.Errorf("failed to update wallet: %w", err)
		}

		// 4. Record lien operation
		if operationType == "lien" {
			_, err = tx.NewInsert().
				Model(lienRecord).
				Exec(ctx)
		} else {
			_, err = tx.NewUpdate().
				Model(lienRecord).
				WherePK().
				Column("released_at").
				Exec(ctx)
		}

		if err != nil {
			return fmt.Errorf("failed to record %s operation: %w", operationType, err)
		}

		return nil
	})
	if err != nil {
		return nil, nil, fmt.Errorf("%s processing failed: %w", operationType, err)
	}

	return lienRecord, wallet, nil
}
