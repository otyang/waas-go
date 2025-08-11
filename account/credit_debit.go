package account

import (
	"context"
	"fmt"

	"github.com/otyang/waas-go/types"
	"github.com/uptrace/bun"
)

// CreditWallet credits funds to a wallet and records the transaction
func (r *WalletRepository) CreditWallet(
	ctx context.Context,
	walletID string,
	creditTx types.CreditTransaction,
) (*types.TransactionHistory, *types.Wallet, error) {
	// 1. Retrieve the wallet with lock to prevent concurrent modifications
	wallet, err := r.FindWalletByID(ctx, walletID)
	if err != nil {
		return nil, nil, err
	}

	// 2. Perform the credit operation
	txHistory, err := wallet.Credit(creditTx)
	if err != nil {
		return txHistory, wallet, err
	}

	// 3. Update wallet and record transaction atomically
	err = r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		if _, err := r.NewWithTx(tx).UpdateWallet(ctx, wallet); err != nil {
			return fmt.Errorf("failed to update wallet balance: %w", err)
		}
		if _, err := r.NewWithTx(tx).CreateTransaction(ctx, txHistory); err != nil {
			return fmt.Errorf("failed to record transaction: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	return txHistory, wallet, nil
}

// DebitWallet debits funds from a wallet and records the transaction
func (r *WalletRepository) DebitWallet(
	ctx context.Context,
	walletID string,
	debitTx types.DebitTransaction,
) (*types.TransactionHistory, *types.Wallet, error) {
	// 1. Retrieve the wallet with lock to prevent concurrent modifications
	wallet, err := r.FindWalletByID(ctx, walletID)
	if err != nil {
		return nil, nil, err
	}

	// 2. Perform the debit operation
	txHistory, err := wallet.Debit(debitTx)
	if err != nil {
		return txHistory, wallet, fmt.Errorf("debit failed: %w", err)
	}

	// 3. Update wallet and record transaction atomically
	err = r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		if _, err := r.NewWithTx(tx).UpdateWallet(ctx, wallet); err != nil {
			return fmt.Errorf("failed to update wallet balance: %w", err)
		}
		if _, err := r.NewWithTx(tx).CreateTransaction(ctx, txHistory); err != nil {
			return fmt.Errorf("failed to record transaction: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	return txHistory, wallet, err
}
