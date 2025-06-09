package types

import (
	"strings"
	"sync"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestNewWallet(t *testing.T) {
	t.Parallel()

	// Arrange
	customerID := "user1"
	currencyCode := "USD"

	// Act
	wallet := NewWallet(customerID, currencyCode)

	// Assert
	assert.NotEmpty(t, wallet.ID)
	assert.Equal(t, customerID, wallet.CustomerID)
	assert.True(t, strings.EqualFold(currencyCode, wallet.CurrencyCode))
	assert.Equal(t, decimal.Zero, wallet.AvailableBalance)
	assert.False(t, wallet.IsClosed)
	assert.False(t, wallet.IsFrozen)
	assert.True(t, wallet.CreatedAt.Before(wallet.UpdatedAt))
}

func TestWallet_TotalBalance(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name             string
		availableBalance decimal.Decimal
		lienBalance      decimal.Decimal
		expectedBalance  decimal.Decimal
	}{
		{
			name:             "Standard balance",
			availableBalance: decimal.NewFromFloat(100.0),
			lienBalance:      decimal.NewFromFloat(20.0),
			expectedBalance:  decimal.NewFromFloat(120.0),
		},
		{
			name:             "Zero available balance",
			availableBalance: decimal.NewFromFloat(0.0),
			lienBalance:      decimal.NewFromFloat(50.0),
			expectedBalance:  decimal.NewFromFloat(50.0),
		},
		{
			name:             "Zero lien balance",
			availableBalance: decimal.NewFromFloat(25.5),
			lienBalance:      decimal.NewFromFloat(0.0),
			expectedBalance:  decimal.NewFromFloat(25.5),
		},
		// Add more test cases as needed
	}

	for _, tc := range testCases {
		tc := tc // Capture range variable to avoid issues
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			wallet := &Wallet{
				AvailableBalance: tc.availableBalance,
				LienBalance:      tc.lienBalance,
			}

			assert.Equal(t, tc.expectedBalance.String(), wallet.TotalBalance().String())
		})
	}
}

func TestWallet_Freeze_Unfreeze(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		initialWallet  *Wallet
		operation      func(w *Wallet) error
		expectedError  error
		expectedWallet *Wallet
	}{
		{
			"Freeze active wallet",
			&Wallet{IsFrozen: false, IsClosed: false},
			func(w *Wallet) error { return w.Freeze() },
			nil,
			&Wallet{IsFrozen: true, IsClosed: false},
		},
		{
			"Unfreeze frozen wallet",
			&Wallet{IsFrozen: true, IsClosed: false},
			func(w *Wallet) error { return w.Unfreeze() },
			nil,
			&Wallet{IsFrozen: false, IsClosed: false},
		},

		{
			"Freeze closed wallet",
			&Wallet{IsFrozen: false, IsClosed: true},
			func(w *Wallet) error { return w.Freeze() },
			ErrWalletClosed,
			&Wallet{IsFrozen: false, IsClosed: true},
		},
		{
			"Unfreeze closed wallet",
			&Wallet{IsFrozen: true, IsClosed: true},
			func(w *Wallet) error { return w.Unfreeze() },
			ErrWalletClosed,
			&Wallet{IsFrozen: true, IsClosed: true},
		},
	}

	for _, tt := range tests {
		tt := tt // Pin the value (https://golang.org/doc/faq#closures_and_goroutines)
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.operation(tt.initialWallet)
			assert.Equal(t, tt.expectedError, err)
			assert.Equal(t, tt.expectedWallet.IsFrozen, tt.initialWallet.IsFrozen)
			assert.Equal(t, tt.expectedWallet.IsClosed, tt.initialWallet.IsClosed)
		})
	}
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
			amount: decimal.NewFromInt(50), fee: decimal.NewFromInt(10),
			expectedErr: nil, expectedBal: decimal.NewFromInt(140),
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
			name: "Closed wallet",
			wallet: &Wallet{
				IsClosed:         true,
				AvailableBalance: decimal.NewFromInt(100),
			},
			amount:      decimal.NewFromInt(50),
			fee:         decimal.NewFromInt(10),
			expectedErr: ErrWalletClosed,
			expectedBal: decimal.NewFromInt(100),
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
			name:                     "Debit closed wallet",
			wallet:                   &Wallet{AvailableBalance: decimal.NewFromInt(100), IsClosed: true},
			amount:                   decimal.NewFromInt(50),
			fee:                      decimal.NewFromInt(10),
			expectedAvailableBalance: decimal.NewFromInt(100),
			expectedError:            ErrWalletClosed,
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
			fromWallet:          NewWallet("owner1", "USD"),
			toWallet:            nil,
			transferAmount:      decimal.NewFromInt(50),
			expectedError:       ErrWalletInvalid,
			expectedFromBalance: decimal.NewFromInt(100),
		},
		{
			name:                "Insufficient transfer total balance",
			fromWallet:          NewWallet("user1", "USD"),
			toWallet:            NewWallet("user2", "USD"),
			transferAmount:      decimal.NewFromInt(500),
			expectedError:       ErrWalletInsufficientBalance,
			expectedFromBalance: decimal.NewFromInt(100),
		},
		{
			name:                "Insufficient transfer to same user",
			fromWallet:          NewWallet("user1", "USD"),
			toWallet:            NewWallet("user1", "USD"),
			transferAmount:      decimal.NewFromInt(500),
			expectedError:       ErrWalletInsufficientBalance,
			expectedFromBalance: decimal.NewFromInt(100),
		},
		{
			name:                "Successful transfer to different user",
			fromWallet:          NewWallet("user1", "USD"),
			toWallet:            NewWallet("user2", "USD"),
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
				assert.Equal(t, tc.expectedFromBalance.String(), tc.fromWallet.AvailableBalance.String())
				assert.Equal(t, tc.expectedToBalance.String(), tc.toWallet.AvailableBalance.String())
			}
		})
	}
}

func TestWallet_TransferTo_Concurrent(t *testing.T) {
	t.Parallel()

	// Create two wallets for the test.
	wallet1 := NewWallet("user1", "USD")
	wallet2 := NewWallet("user2", "USD")

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
		wallet1 := NewWallet("user1", "USD")
		wallet1.AvailableBalance = decimal.NewFromFloat(100)

		wallet2 := NewWallet("user1", "EUR")
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
		wallet1 := NewWallet("user1", "USD")
		wallet2 := NewWallet("user1", "USD")

		err := wallet1.Swap(wallet2, decimal.Zero, decimal.Zero, decimal.Zero)
		assert.Equal(t, ErrWalletSameCurrencySwap, err)
	})

	t.Run("SwapMustBeSameOwner", func(t *testing.T) {
		wallet1 := NewWallet("user1", "USD")
		wallet2 := NewWallet("user2", "EUR")

		err := wallet1.Swap(wallet2, decimal.Zero, decimal.Zero, decimal.Zero)
		assert.Equal(t, ErrWalletSwapSameOwnerRequired, err)
	})

	t.Run("InsufficientBalance", func(t *testing.T) {
		wallet1 := NewWallet("user1", "USD")
		wallet2 := NewWallet("user1", "EUR")

		err := wallet1.Swap(wallet2, decimal.NewFromFloat(100), decimal.Zero, decimal.Zero)
		assert.Equal(t, ErrWalletInsufficientBalance, err)
	})

	t.Run("Swap TO A closed or frozen account", func(t *testing.T) {
		wallet1 := NewWallet("user1", "USD")
		wallet1.AvailableBalance = decimal.NewFromFloat(100)

		wallet2 := NewWallet("user1", "EUR")
		wallet2.IsClosed = true

		err := wallet1.Swap(wallet2, decimal.NewFromFloat(100), decimal.Zero, decimal.Zero)
		assert.Equal(t, ErrWalletClosed.Error(), err.Error())

		wallet2.IsFrozen = true
		err = wallet1.Swap(wallet2, wallet1.AvailableBalance, decimal.NewFromFloat(40), decimal.Zero)
		assert.Error(t, err)
		assert.Equal(t, decimal.NewFromFloat(0).String(), wallet2.AvailableBalance.String())
	})

	t.Run("Swap FROM A closed or frozen account", func(t *testing.T) {
		wallet1 := NewWallet("user1", "USD")
		wallet1.AvailableBalance = decimal.NewFromFloat(100)
		wallet1.IsClosed = true

		wallet2 := NewWallet("user1", "EUR")

		err := wallet1.Swap(wallet2, decimal.NewFromFloat(100), decimal.Zero, decimal.Zero)
		assert.Equal(t, ErrWalletClosed.Error(), err.Error())

		wallet1.IsFrozen = true
		err = wallet1.Swap(wallet2, decimal.NewFromFloat(100), decimal.Zero, decimal.Zero)
		assert.Equal(t, ErrWalletFrozen.Error(), err.Error())
	})
}

func TestWallet_LienAmount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		initialBalance  decimal.Decimal
		lockAmount      decimal.Decimal
		expectedBalance decimal.Decimal
		expectedError   error
	}{
		{
			name:            "LockAmount Exceeds Available Balance",
			initialBalance:  decimal.NewFromInt(100),
			lockAmount:      decimal.NewFromInt(200),
			expectedBalance: decimal.NewFromInt(100),
			expectedError:   ErrWalletInsufficientBalance,
		},
		{
			name:            "Lock Remaining Balance",
			initialBalance:  decimal.NewFromInt(100),
			lockAmount:      decimal.NewFromInt(100),
			expectedBalance: decimal.Zero,
			expectedError:   nil,
		},
		{
			name:            "LockAmount Insufficient Funds FrozenWallet",
			initialBalance:  decimal.NewFromInt(100),
			lockAmount:      decimal.NewFromInt(200),
			expectedBalance: decimal.NewFromInt(100),
			expectedError:   ErrWalletInsufficientBalance,
		},
		{
			name:            "Lock Negative Amount",
			initialBalance:  decimal.NewFromInt(100),
			lockAmount:      decimal.NewFromInt(-10),
			expectedBalance: decimal.NewFromInt(100),
			expectedError:   ErrWalletInvalidAmount,
		},
		{
			name:            "Closed Wallet",
			initialBalance:  decimal.NewFromInt(100),
			lockAmount:      decimal.NewFromInt(50),
			expectedBalance: decimal.NewFromInt(100),
			expectedError:   ErrWalletClosed,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			wallet := &Wallet{
				AvailableBalance: tc.initialBalance,
				// Status:           WalletStatusFrozen,
			}

			if tc.name == "Closed Wallet" {
				wallet.IsClosed = true
			}

			err := wallet.LienAmount(tc.lockAmount)

			assert.Equal(t, tc.expectedError, err)
			assert.True(t, tc.expectedBalance.Equal(wallet.AvailableBalance))
		})
	}
}

func TestWallet_LienAmount_Concurrent(t *testing.T) {
	t.Parallel()

	wallet := &Wallet{}
	wallet.AvailableBalance = decimal.NewFromInt(1000)

	var wg sync.WaitGroup

	// Run multiple concurrent lock requests for the same amount.
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			err := wallet.LienAmount(decimal.NewFromInt(50))

			if i < 2 {
				assert.NoError(t, err)
			}
		}(i)
	}

	wg.Wait()

	// Verify the final locked balance is equal to the sum of requested amounts.
	expectedLienBalance := decimal.NewFromInt(500)
	assert.True(t, expectedLienBalance.Equal(wallet.LienBalance))
}

// file too long see next file for second  test
func TestWallet_UnlienAmount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		wallet          *Wallet
		amount          decimal.Decimal
		expectedBalance decimal.Decimal
		expectedError   error
	}{
		{
			name: "unlock positive amount",
			wallet: &Wallet{
				AvailableBalance: decimal.NewFromInt(100),
				LienBalance:      decimal.NewFromInt(50),
			},
			amount:          decimal.NewFromInt(20),
			expectedBalance: decimal.NewFromInt(120),
			expectedError:   nil,
		},
		{
			name: "unlock negative amount",
			wallet: &Wallet{
				AvailableBalance: decimal.NewFromInt(100),
				LienBalance:      decimal.NewFromInt(50),
			},
			amount:          decimal.NewFromInt(-10),
			expectedBalance: decimal.NewFromInt(100),
			expectedError:   ErrWalletInvalidAmount,
		},
		{
			name: "unlock insufficient amount",
			wallet: &Wallet{
				AvailableBalance: decimal.NewFromInt(100),
				LienBalance:      decimal.NewFromInt(10),
			},
			amount:          decimal.NewFromInt(20),
			expectedBalance: decimal.NewFromInt(100),
			expectedError:   ErrWalletInsufficientBalance,
		},
		{
			name: "unlocking when locked balance exceeds available balance",
			wallet: &Wallet{
				LienBalance:      decimal.NewFromInt(10),
				AvailableBalance: decimal.NewFromInt(5),
			},
			amount:          decimal.NewFromInt(15),
			expectedBalance: decimal.NewFromInt(5),
			expectedError:   ErrWalletInsufficientBalance,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.wallet.UnLienAmount(tc.amount)
			assert.Equal(t, tc.expectedError, err)
			assert.Equal(t, tc.expectedBalance, tc.wallet.AvailableBalance)
		})
	}
}
