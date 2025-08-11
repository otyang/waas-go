package store

import (
	"context"
	"fmt"

	"github.com/otyang/waas-go/types"
	"github.com/uptrace/bun"
)

// FreezeWallet freezes a wallet and records the freeze operation
func (r *WalletRepository) FreezeWallet(
	ctx context.Context,
	walletID string,
	req types.FreezeRequest,
) (*types.FreezeResult, *types.Wallet, error) {
	// Validate request
	if req.Reason == "" {
		return nil, nil, fmt.Errorf("freeze reason is required")
	}
	if req.InitiatedBy == "" {
		return nil, nil, fmt.Errorf("initiator is required")
	}

	var (
		freezeResult *types.FreezeResult
		wallet       *types.Wallet
	)

	err := r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		// Get repository with transaction
		repo := r.NewWithTx(tx)

		// 1. Retrieve wallet
		var err error
		wallet, err = repo.FindWalletByID(ctx, walletID)
		if err != nil {
			return fmt.Errorf("failed to get wallet: %w", err)
		}

		// 2. Perform the freeze operation
		freezeResult, err = wallet.Freeze(req)
		if err != nil {
			return fmt.Errorf("freeze operation failed: %w", err)
		}

		// 3. Update wallet
		if _, err := repo.UpdateWallet(ctx, wallet); err != nil {
			return fmt.Errorf("failed to update wallet: %w", err)
		}

		// 4. Record freeze operation
		if _, err := tx.NewInsert().
			Model(freezeResult).
			Exec(ctx); err != nil {
			return fmt.Errorf("failed to record freeze: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, nil, fmt.Errorf("freeze processing failed: %w", err)
	}

	return freezeResult, wallet, nil
}

// UnfreezeWallet unfreezes a wallet and records the operation
func (r *WalletRepository) UnfreezeWallet(
	ctx context.Context,
	walletID string,
	req types.UnfreezeRequest,
) (*types.UnfreezeResult, *types.Wallet, error) {
	// Validate request
	if req.Reason == "" {
		return nil, nil, fmt.Errorf("unfreeze reason is required")
	}
	if req.InitiatedBy == "" {
		return nil, nil, fmt.Errorf("initiator is required")
	}

	var (
		unfreezeResult *types.UnfreezeResult
		wallet         *types.Wallet
	)

	err := r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		// Get repository with transaction
		repo := r.NewWithTx(tx)

		// 1. Retrieve wallet
		var err error
		wallet, err = repo.FindWalletByID(ctx, walletID)
		if err != nil {
			return fmt.Errorf("failed to get wallet: %w", err)
		}

		// 2. Perform the unfreeze operation
		unfreezeResult, err = wallet.Unfreeze(req)
		if err != nil {
			return fmt.Errorf("unfreeze operation failed: %w", err)
		}

		// 3. Update wallet
		if _, err := repo.UpdateWallet(ctx, wallet); err != nil {
			return fmt.Errorf("failed to update wallet: %w", err)
		}

		// 4. Record unfreeze operation
		if _, err := tx.NewInsert().
			Model(unfreezeResult).
			Exec(ctx); err != nil {
			return fmt.Errorf("failed to record unfreeze: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, nil, fmt.Errorf("unfreeze processing failed: %w", err)
	}

	return unfreezeResult, wallet, nil
}

// GetFreezeStatus returns current freeze status of a wallet
func (r *WalletRepository) GetFreezeStatus(ctx context.Context, walletID string) (*types.FreezeResult, error) {
	wallet, err := r.FindWalletByID(ctx, walletID)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}

	return wallet.GetFreezeInfo(), nil
}
