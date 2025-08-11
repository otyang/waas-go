package store

import (
	"context"
	"fmt"

	"github.com/otyang/waas-go/types"
	"github.com/shopspring/decimal"
	"github.com/uptrace/bun"
)

// TransferFunds transfers money between wallets and records both transactions atomically
func (r *WalletRepository) TransferFunds(
	ctx context.Context,
	sourceWalletID string,
	destWalletID string,
	req types.TransferRequest,
) (*types.TransactionHistory, *types.TransactionHistory, error) {
	// Validate basic request parameters
	if req.Amount.LessThanOrEqual(decimal.Zero) {
		return nil, nil, types.ErrInvalidAmount
	}
	if req.Fee.LessThan(decimal.Zero) {
		return nil, nil, types.ErrInvalidFee
	}

	var (
		sourceTx, destTx         *types.TransactionHistory
		sourceWallet, destWallet *types.Wallet
	)

	// Execute in transaction
	err := r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		// Get repositories with transaction
		theRepo := r.NewWithTx(tx)

		// 1. Retrieve both wallets with locking
		var err error
		sourceWallet, err = theRepo.FindWalletByID(ctx, sourceWalletID)
		if err != nil {
			return fmt.Errorf("failed to get source wallet: %w", err)
		}

		destWallet, err = theRepo.FindWalletByID(ctx, destWalletID)
		if err != nil {
			return fmt.Errorf("failed to get destination wallet: %w", err)
		}

		// 2. Perform the transfer
		sourceTx, destTx, err = sourceWallet.Transfer(destWallet, req)
		if err != nil {
			return fmt.Errorf("transfer validation failed: %w", err)
		}

		// 3. Update both wallets
		if _, err := theRepo.UpdateWallet(ctx, sourceWallet); err != nil {
			return fmt.Errorf("failed to update source wallet: %w", err)
		}
		if _, err := theRepo.UpdateWallet(ctx, destWallet); err != nil {
			return fmt.Errorf("failed to update destination wallet: %w", err)
		}

		// 4. Record both transactions
		if _, err := theRepo.CreateTransaction(ctx, sourceTx); err != nil {
			return fmt.Errorf("failed to record source transaction: %w", err)
		}
		if _, err := theRepo.CreateTransaction(ctx, destTx); err != nil {
			return fmt.Errorf("failed to record destination transaction: %w", err)
		}

		return nil
	})
	if err != nil {
		// Return the transaction records even if failed (they contain failure status)
		if sourceTx != nil && destTx != nil {
			return sourceTx, destTx, fmt.Errorf("transfer failed: %w", err)
		}
		return nil, nil, fmt.Errorf("transfer failed before execution: %w", err)
	}

	return sourceTx, destTx, nil
}

// SwapFunds exchanges funds between wallets of different currencies at a specified rate
func (r *WalletRepository) SwapFunds(
	ctx context.Context,
	sourceWalletID string,
	destWalletID string,
	req types.SwapRequest,
) (*types.TransactionHistory, *types.TransactionHistory, error) {
	// Validate basic request parameters
	if req.SourceAmount.LessThanOrEqual(decimal.Zero) {
		return nil, nil, types.ErrInvalidAmount
	}
	if req.DestinationAmount.LessThanOrEqual(decimal.Zero) {
		return nil, nil, types.ErrInvalidAmount
	}
	if req.Fee.LessThan(decimal.Zero) {
		return nil, nil, types.ErrInvalidFee
	}
	if req.ExchangeRate.LessThanOrEqual(decimal.Zero) {
		return nil, nil, types.ErrInvalidExchangeRate
	}

	var (
		sourceTx, destTx         *types.TransactionHistory
		sourceWallet, destWallet *types.Wallet
	)

	// Execute in transaction
	err := r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		// Get repositories with transaction
		theRepo := r.NewWithTx(tx)

		// 1. Retrieve both wallets with locking
		var err error
		sourceWallet, err = theRepo.FindWalletByID(ctx, sourceWalletID)
		if err != nil {
			return fmt.Errorf("failed to get source wallet: %w", err)
		}

		destWallet, err = theRepo.FindWalletByID(ctx, destWalletID)
		if err != nil {
			return fmt.Errorf("failed to get destination wallet: %w", err)
		}

		// 2. Verify exchange rate matches the amounts
		expectedDestAmount := req.SourceAmount.Mul(req.ExchangeRate)
		if !req.DestinationAmount.Equal(expectedDestAmount) {
			return fmt.Errorf("%w: calculated amount %s doesn't match provided amount %s",
				types.ErrExchangeRateMismatch,
				expectedDestAmount.String(),
				req.DestinationAmount.String(),
			)
		}

		// 3. Perform the swap
		sourceTx, destTx, err = sourceWallet.Swap(destWallet, req)
		if err != nil {
			return fmt.Errorf("swap validation failed: %w", err)
		}

		// 4. Update both wallets
		if _, err := theRepo.UpdateWallet(ctx, sourceWallet); err != nil {
			return fmt.Errorf("failed to update source wallet: %w", err)
		}
		if _, err := theRepo.UpdateWallet(ctx, destWallet); err != nil {
			return fmt.Errorf("failed to update destination wallet: %w", err)
		}

		// 5. Record both transactions
		if _, err := theRepo.CreateTransaction(ctx, sourceTx); err != nil {
			return fmt.Errorf("failed to record source transaction: %w", err)
		}
		if _, err := theRepo.CreateTransaction(ctx, destTx); err != nil {
			return fmt.Errorf("failed to record destination transaction: %w", err)
		}

		return nil
	})
	if err != nil {
		// Return the transaction records even if failed (they contain failure status)
		if sourceTx != nil && destTx != nil {
			return sourceTx, destTx, fmt.Errorf("swap failed: %w", err)
		}
		return nil, nil, fmt.Errorf("swap failed before execution: %w", err)
	}

	return sourceTx, destTx, nil
}
