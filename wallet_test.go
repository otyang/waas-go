package waas

import (
	"sync"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestGenerateWalletID(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		currencyCode string
		userID       string
		expectedID   string
	}{
		{name: "Valid inputs", currencyCode: "USD", userID: "user1-Nick", expectedID: "usd-user1-Nick"},
		{name: "Currency code in lower case", currencyCode: "eur", userID: "user2", expectedID: "eur-user2"},
		{name: "Currency code in upper case", currencyCode: "BTC", userID: "user-3", expectedID: "btc-user-3"},
		{name: "Special characters in user ID", currencyCode: "BTC", userID: "user_4!", expectedID: "btc-user_4!"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualID := GenerateWalletID(tc.currencyCode, tc.userID)
			assert.Equal(t, tc.expectedID, actualID)
		})
	}
}

func TestNewWallet(t *testing.T) {
	t.Parallel()

	// Arrange
	customerID := "user1"
	currencyCode := "USD"
	isFiat := true

	// Act
	wallet := NewWallet(customerID, currencyCode, isFiat)

	// Assert
	assert.Equal(t, GenerateWalletID(currencyCode, customerID), wallet.ID)
	assert.Equal(t, customerID, wallet.CustomerID)
	assert.Equal(t, currencyCode, wallet.CurrencyCode)
	assert.Equal(t, decimal.Zero, wallet.AvailableBalance)
	assert.False(t, wallet.IsFrozen)
	assert.True(t, wallet.IsFiat)
	assert.True(t, wallet.CreatedAt.Before(wallet.UpdatedAt))
}

func TestWallet_TotalBalance(t *testing.T) {
	t.Parallel()

	// Arrange
	wallet := &Wallet{}
	wallet.AvailableBalance = decimal.NewFromFloat(100.0)

	// Act
	totalBalance := wallet.TotalBalance()
	assert.Equal(t, decimal.NewFromFloat(100.0), totalBalance)
}

func TestWallet_Freeze_Unfreeze(t *testing.T) {
	t.Parallel()

	wallet := &Wallet{}

	wallet.Freeze()
	assert.True(t, wallet.IsFrozen)

	wallet.Unfreeze()
	assert.False(t, wallet.IsFrozen)
}

func TestCreditBalance(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		wallet      *Wallet
		amount      decimal.Decimal
		fee         decimal.Decimal
		expectedErr error
		expectedBal decimal.Decimal
	}{
		{
			name: "Success",
			wallet: &Wallet{
				AvailableBalance: decimal.NewFromInt(100),
			},
			amount:      decimal.NewFromInt(50),
			fee:         decimal.NewFromInt(10),
			expectedErr: nil,
			expectedBal: decimal.NewFromInt(140),
		},
		{
			name: "Frozen wallet",
			wallet: &Wallet{
				IsFrozen:         true,
				AvailableBalance: decimal.NewFromInt(100),
			},
			amount:      decimal.NewFromInt(50),
			fee:         decimal.NewFromInt(10),
			expectedErr: nil,
			expectedBal: decimal.NewFromInt(140),
		},
		{
			name: "Invalid amount",
			wallet: &Wallet{
				AvailableBalance: decimal.NewFromInt(100),
			},
			amount:      decimal.Zero,
			fee:         decimal.NewFromInt(10),
			expectedErr: ErrWalletInvalidAmount,
			expectedBal: decimal.NewFromInt(100),
		},
		{
			name: "Insufficient funds",
			wallet: &Wallet{
				AvailableBalance: decimal.NewFromInt(10),
			},
			amount:      decimal.NewFromInt(20),
			fee:         decimal.NewFromInt(10),
			expectedErr: nil,
			expectedBal: decimal.NewFromInt(20),
		},
		{
			name: "Crediting with negative fee",
			wallet: &Wallet{
				AvailableBalance: decimal.Zero,
			},
			amount:      decimal.NewFromInt(100),
			fee:         decimal.NewFromInt(-10),
			expectedErr: ErrWalletInvalidAmount,
			expectedBal: decimal.Zero,
		},
		{
			name: "Crediting with negative amount",
			wallet: &Wallet{
				AvailableBalance: decimal.NewFromInt(10),
			},
			amount:      decimal.NewFromInt(-100),
			fee:         decimal.NewFromInt(10),
			expectedErr: ErrWalletInvalidAmount,
			expectedBal: decimal.NewFromInt(10),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.wallet.CreditBalance(tc.amount, tc.fee)
			assert.Equal(t, tc.expectedErr, err)
			assert.Equal(t, tc.expectedBal, tc.wallet.AvailableBalance)
		})
	}
}

func TestDebitBalance(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                     string
		wallet                   *Wallet
		amount                   decimal.Decimal
		fee                      decimal.Decimal
		expectedAvailableBalance decimal.Decimal
		expectedError            error
	}{
		{
			name:                     "Debit positive amount and positive fee",
			wallet:                   &Wallet{AvailableBalance: decimal.NewFromInt(100)},
			amount:                   decimal.NewFromInt(50),
			fee:                      decimal.NewFromInt(10),
			expectedAvailableBalance: decimal.NewFromInt(40),
			expectedError:            nil,
		},
		{
			name:                     "Debit frozen wallet",
			wallet:                   &Wallet{AvailableBalance: decimal.NewFromInt(100), IsFrozen: true},
			amount:                   decimal.NewFromInt(50),
			fee:                      decimal.NewFromInt(10),
			expectedAvailableBalance: decimal.NewFromInt(100),
			expectedError:            ErrWalletFrozen,
		},
		{
			name:                     "Debit negative amount",
			wallet:                   &Wallet{AvailableBalance: decimal.NewFromInt(100)},
			amount:                   decimal.NewFromInt(-50),
			fee:                      decimal.NewFromInt(10),
			expectedAvailableBalance: decimal.NewFromInt(100),
			expectedError:            ErrWalletInvalidAmount,
		},
		{
			name:                     "Debit negative fee",
			wallet:                   &Wallet{AvailableBalance: decimal.NewFromInt(100)},
			amount:                   decimal.NewFromInt(50),
			fee:                      decimal.NewFromInt(-10),
			expectedAvailableBalance: decimal.NewFromInt(100),
			expectedError:            ErrWalletInvalidAmount,
		},
		{
			name:                     "Insufficient funds",
			wallet:                   &Wallet{AvailableBalance: decimal.NewFromInt(10)},
			amount:                   decimal.NewFromInt(20),
			fee:                      decimal.NewFromInt(10),
			expectedAvailableBalance: decimal.NewFromInt(10),
			expectedError:            ErrWalletInsufficientBalance,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.wallet.DebitBalance(tc.amount, tc.fee)
			assert.Equal(t, tc.expectedError, err)
			assert.Equal(t, tc.expectedAvailableBalance, tc.wallet.AvailableBalance)
		})
	}
}

func TestConcurrentCreditDebitBalance(t *testing.T) {
	t.Parallel()

	wallet := &Wallet{
		AvailableBalance: decimal.NewFromInt(100),
		IsFrozen:         false,
	}

	var wg sync.WaitGroup
	wg.Add(2)

	{
		creditAmount := decimal.NewFromInt(50)
		creditFee := decimal.NewFromInt(10)

		go func() {
			defer wg.Done()
			err := wallet.CreditBalance(creditAmount, creditFee)
			if err != nil {
				t.Errorf("CreditBalance error: %v", err)
			}
		}()
		go func() {
			defer wg.Done()
			err := wallet.CreditBalance(creditAmount, creditFee)
			if err != nil {
				t.Errorf("CreditBalance error: %v", err)
			}
		}()
	}
	wg.Wait()

	{
		debitAmount := decimal.NewFromInt(120)
		debitFee := decimal.NewFromInt(5)

		err := wallet.DebitBalance(debitAmount, debitFee)
		assert.NoError(t, err)
	}

	assert.Equal(t, decimal.NewFromInt(55), wallet.AvailableBalance)
}

func TestWallet_Transfer(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                                   string
		fromWallet, toWallet                   *Wallet
		transferAmount                         decimal.Decimal
		fee                                    decimal.Decimal
		expectedError                          error
		expectedFromBalance, expectedToBalance decimal.Decimal
	}{
		{
			name:                "Transfer with nil ToWallet",
			fromWallet:          NewWallet("owner1", "USD", true),
			toWallet:            nil,
			transferAmount:      decimal.NewFromInt(50),
			expectedError:       ErrWalletInvalid,
			expectedFromBalance: decimal.NewFromInt(100),
		},
		{
			name:                "Insufficient transfer total balance",
			fromWallet:          NewWallet("user1", "USD", true),
			toWallet:            NewWallet("user2", "USD", true),
			transferAmount:      decimal.NewFromInt(500),
			expectedError:       ErrWalletInsufficientBalance,
			expectedFromBalance: decimal.NewFromInt(100),
		},
		{
			name:                "Insufficient transfer to same user",
			fromWallet:          NewWallet("user1", "USD", true),
			toWallet:            NewWallet("user1", "EUR", true),
			transferAmount:      decimal.NewFromInt(500),
			expectedError:       ErrWalletInsufficientBalance,
			expectedFromBalance: decimal.NewFromInt(100),
		},
		{
			name:                "Successful transfer to different user",
			fromWallet:          NewWallet("user1", "USD", true),
			toWallet:            NewWallet("user2", "USD", true),
			transferAmount:      decimal.NewFromInt(50),
			fee:                 decimal.NewFromInt(1),
			expectedError:       nil,
			expectedFromBalance: decimal.NewFromInt(49),
			expectedToBalance:   decimal.NewFromInt(50),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.fromWallet.AvailableBalance = decimal.NewFromInt(100)
			err := tc.fromWallet.TransferTo(tc.toWallet, tc.transferAmount, tc.fee)
			assert.Equal(t, tc.expectedError, err)
			if tc.expectedError == nil {
				assert.True(t, tc.expectedFromBalance.Equal(tc.fromWallet.AvailableBalance))
				assert.True(t, tc.expectedToBalance.Equal(tc.toWallet.AvailableBalance))
			}
		})
	}
}

func TestWallet_TransferTo_Concurrent(t *testing.T) {
	t.Parallel()

	// Create two wallets for the test.
	wallet1 := NewWallet("user1", "USD", true)
	wallet2 := NewWallet("user2", "USD", true)

	// Initialize wallets with some balance.
	wallet1.AvailableBalance = decimal.NewFromInt(100)
	wallet2.AvailableBalance = decimal.NewFromInt(50)

	// Define the transfer amount and expected outcome.
	transferAmount := decimal.NewFromInt(25)

	// Define two goroutines for concurrent transfer attempts.
	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		err := wallet1.TransferTo(wallet2, transferAmount, decimal.Zero)
		assert.NoError(t, err)
	}()

	go func() {
		defer wg.Done()
		err := wallet2.TransferTo(wallet1, transferAmount, decimal.Zero)
		assert.NoError(t, err)
	}()

	// Wait for both goroutines to finish.
	wg.Wait()

	// Assert that the balances of both wallets are updated correctly.
	assert.Equal(t, decimal.NewFromInt(100), wallet1.AvailableBalance)
	assert.Equal(t, decimal.NewFromInt(50), wallet2.AvailableBalance)
}

func TestSwap(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		wallet1 := NewWallet("user1", "USD", true)
		wallet1.AvailableBalance = decimal.NewFromFloat(100)

		wallet2 := NewWallet("user1", "EUR", true)
		wallet2.AvailableBalance = decimal.NewFromFloat(50)

		fromAmount := decimal.NewFromFloat(50)
		toAmount := decimal.NewFromFloat(40)
		fee := decimal.NewFromFloat(1)

		err := wallet1.Swap(wallet2, fromAmount, toAmount, fee)
		assert.NoError(t, err)
		assert.Equal(t, decimal.NewFromFloat(49), wallet1.AvailableBalance)
		assert.Equal(t, decimal.NewFromFloat(90), wallet2.AvailableBalance)
	})

	t.Run("InvalidWallet", func(t *testing.T) {
		err := (&Wallet{}).Swap(nil, decimal.Zero, decimal.Zero, decimal.Zero)
		assert.Equal(t, ErrWalletInvalid, err)
	})

	t.Run("SameCurrencySwap", func(t *testing.T) {
		wallet1 := NewWallet("user1", "USD", true)
		wallet2 := NewWallet("user1", "USD", true)

		err := wallet1.Swap(wallet2, decimal.Zero, decimal.Zero, decimal.Zero)
		assert.Equal(t, ErrWalletSameCurrencySwap, err)
	})

	t.Run("SwapMustBeSameOwner", func(t *testing.T) {
		wallet1 := NewWallet("user1", "USD", true)
		wallet2 := NewWallet("user2", "EUR", true)

		err := wallet1.Swap(wallet2, decimal.Zero, decimal.Zero, decimal.Zero)
		assert.Equal(t, ErrWalletSwapSameOwnerRequired, err)
	})

	t.Run("InsufficientBalance", func(t *testing.T) {
		wallet1 := NewWallet("user1", "USD", true)
		wallet2 := NewWallet("user1", "EUR", true)

		err := wallet1.Swap(wallet2, decimal.NewFromFloat(100), decimal.Zero, decimal.Zero)
		assert.Equal(t, ErrWalletInsufficientBalance, err)
	})
}
