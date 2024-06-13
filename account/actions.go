package account

// func (a *Client) Credit(ctx context.Context, opts types.CreditWalletOpts) (*types.CreditWalletResponse, error) {
// 	wallet, err := a.GetWalletByID(ctx, opts.WalletID)
// 	if err != nil {
// 		return nil, err
// 	}

// 	err = wallet.CreditBalance(opts.Amount, opts.Fee)
// 	if err != nil {
// 		return nil, err
// 	}

// 	transaction := types.NewTransactionForCreditEntry(wallet, opts.Amount, opts.Fee, opts.Type)
// 	transaction.SetNarration(opts.Narration)

// 	err = a.UpdateWalletAndInsertTransaction(ctx, []*types.Wallet{wallet}, []*types.Transaction{transaction})
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &types.CreditWalletResponse{Wallet: wallet, Transaction: transaction}, nil
// }

// func (a *Client) Debit(ctx context.Context, opts types.DebitWalletOpts) (*types.DebitWalletResponse, error) {
// 	wallet, err := a.GetWalletByID(ctx, opts.WalletID)
// 	if err != nil {
// 		return nil, err
// 	}

// 	err = wallet.DebitBalance(opts.Amount, opts.Fee)
// 	if err != nil {
// 		return nil, err
// 	}

// 	transaction := types.NewTransactionForDebitEntry(wallet, opts.Amount, opts.Fee, opts.Type, opts.Status)
// 	transaction.SetNarration(opts.Narration)

// 	err = a.WithTxBulkUpdateWalletAndInsertTransaction(ctx, []*types.Wallet{wallet}, []*types.Transaction{transaction})
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &types.DebitWalletResponse{Wallet: wallet, Transaction: transaction}, nil
// }

// func (a *Client) Reverse(ctx context.Context, transactionID string) (*types.ReverseResponse, error) {
// 	t, err := a.GetTransaction(ctx, transactionID)
// 	if err != nil {
// 		return nil, err
// 	}

// 	wallet, err := a.GetWalletByID(ctx, t.WalletID)
// 	if err != nil {
// 		return nil, err
// 	}

// 	rr, err := t.Reverse(wallet)
// 	if err != nil {
// 		return nil, err
// 	}
// 	// update transaction status
// 	err = a.WithTxBulkUpdateWalletAndInsertTransaction(ctx, []*types.Wallet{rr.Wallet}, []*types.Transaction{rr.OldTx, rr.NewTx})
// 	if err != nil {
// 		return nil, err
// 	}

// 	return rr, nil
// }
