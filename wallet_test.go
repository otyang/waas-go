package waas

import (
	"strings"
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
	assert.True(t, strings.EqualFold(currencyCode, wallet.CurrencyCode))
	assert.Equal(t, decimal.Zero, wallet.AvailableBalance)
	assert.Equal(t, WalletStatusActive, wallet.Status)
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
		initialStatus  WalletStatus
		operation      func(w *Wallet) error
		expectedError  error
		expectedStatus WalletStatus
	}{
		{"Freeze active wallet", WalletStatusActive, func(w *Wallet) error { return w.Freeze() }, nil, WalletStatusFrozen},
		{"Unfreeze frozen wallet", WalletStatusFrozen, func(w *Wallet) error { return w.Unfreeze() }, nil, WalletStatusActive},
		{"Freeze closed wallet", WalletStatusClosed, func(w *Wallet) error { return w.Freeze() }, ErrWalletClosed, WalletStatusClosed},
		{"Unfreeze closed wallet", WalletStatusClosed, func(w *Wallet) error { return w.Unfreeze() }, ErrWalletClosed, WalletStatusClosed},
	}

	for _, tt := range tests {
		tt := tt // Pin the value (https://golang.org/doc/faq#closures_and_goroutines)
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			wallet := &Wallet{Status: tt.initialStatus}

			err := tt.operation(wallet)
			assert.Equal(t, tt.expectedError, err)
			assert.Equal(t, tt.expectedStatus, wallet.Status)
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
				Status:           WalletStatusFrozen,
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
				Status:           WalletStatusClosed,
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
			wallet:                   &Wallet{AvailableBalance: decimal.NewFromInt(100), Status: WalletStatusFrozen},
			amount:                   decimal.NewFromInt(50),
			fee:                      decimal.NewFromInt(10),
			expectedAvailableBalance: decimal.NewFromInt(100),
			expectedError:            ErrWalletFrozen,
		},
		{
			name:                     "Debit closed wallet",
			wallet:                   &Wallet{AvailableBalance: decimal.NewFromInt(100), Status: WalletStatusClosed},
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
		Status:           WalletStatusActive,
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

func TestWallet_Transfesr(t *testing.T) {
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
				assert.Equal(t, tc.expectedFromBalance.String(), tc.fromWallet.AvailableBalance.String())
				assert.Equal(t, tc.expectedToBalance.String(), tc.toWallet.AvailableBalance.String())
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

	t.Run("Swap TO A closed or frozen account", func(t *testing.T) {
		wallet1 := NewWallet("user1", "USD", true)
		wallet1.AvailableBalance = decimal.NewFromFloat(100)

		wallet2 := NewWallet("user1", "EUR", true)
		wallet2.Status = WalletStatusClosed

		err := wallet1.Swap(wallet2, decimal.NewFromFloat(100), decimal.Zero, decimal.Zero)
		assert.Equal(t, ErrWalletClosed.Error(), err.Error())

		wallet2.Status = WalletStatusFrozen
		err = wallet1.Swap(wallet2, wallet1.AvailableBalance, decimal.NewFromFloat(40), decimal.Zero)
		assert.NoError(t, err)
		assert.Equal(t, wallet2.AvailableBalance.String(), decimal.NewFromFloat(40).String())
	})

	t.Run("Swap FROM A closed or frozen account", func(t *testing.T) {
		wallet1 := NewWallet("user1", "USD", true)
		wallet1.AvailableBalance = decimal.NewFromFloat(100)
		wallet1.Status = WalletStatusClosed

		wallet2 := NewWallet("user1", "EUR", true)

		err := wallet1.Swap(wallet2, decimal.NewFromFloat(100), decimal.Zero, decimal.Zero)
		assert.Equal(t, ErrWalletClosed.Error(), err.Error())

		wallet1.Status = WalletStatusFrozen
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
			name:            "Frozen Wallet",
			initialBalance:  decimal.NewFromInt(100),
			lockAmount:      decimal.NewFromInt(50),
			expectedBalance: decimal.NewFromInt(100),
			expectedError:   ErrWalletFrozen,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			wallet := &Wallet{
				AvailableBalance: tc.initialBalance,
				// Status:           WalletStatusFrozen,
			}

			if tc.name == "Frozen Wallet" {
				wallet.Status = WalletStatusFrozen
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
