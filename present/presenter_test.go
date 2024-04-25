package present

import (
	"testing"

	"github.com/otyang/waas-go/types"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

// Test for success: Verifies correct calculation and response for valid wallets.
func TestWalletListSuccess(t *testing.T) {
	t.Parallel()

	currencies := []types.Currency{
		{
			Code:      "ngn",
			RateBuy:   decimal.NewFromFloat(415),
			Precision: 2,
		},
		{
			Code:      "usd",
			RateBuy:   decimal.NewFromFloat(1),
			Precision: 2,
		},
	}
	// Arrange
	wallets := []*types.Wallet{
		{
			ID:               "1",
			CustomerID:       "customer1",
			CurrencyCode:     "NGN",
			AvailableBalance: decimal.NewFromFloat(1000),
			LienBalance:      decimal.NewFromFloat(0),
		},
		{
			ID:               "2",
			CustomerID:       "customer2",
			CurrencyCode:     "USD",
			AvailableBalance: decimal.NewFromFloat(50),
			LienBalance:      decimal.NewFromFloat(10),
		},
	}

	responses, err := WalletList(wallets, currencies)
	assert.NoError(t, err)

	expectedNewWalletResponses := []NewWalletResponse{
		{
			ID:                "1",
			CustomerID:        "customer1",
			Currency:          currencies[0],
			AvailableBalance:  decimal.NewFromFloat(1000).RoundBank(int32(currencies[0].Precision)),
			LienBalance:       decimal.NewFromFloat(0).RoundBank(int32(currencies[0].Precision)),
			TotalBalance:      decimal.NewFromFloat(1000).RoundBank(int32(currencies[0].Precision)),
			TotalBalanceInUSD: decimal.NewFromFloat(2.41).RoundBank(int32(currencies[0].Precision)),
			IsFrozen:          false,
			IsClosed:          false,
		},
		{
			ID:                "2",
			CustomerID:        "customer2",
			Currency:          currencies[1],
			AvailableBalance:  decimal.NewFromFloat(50).RoundBank(int32(currencies[1].Precision)),
			LienBalance:       decimal.NewFromFloat(10).RoundBank(int32(currencies[1].Precision)),
			TotalBalance:      decimal.NewFromFloat(60).RoundBank(int32(currencies[1].Precision)),
			TotalBalanceInUSD: decimal.NewFromFloat(60).RoundBank(int32(currencies[1].Precision)),
			IsFrozen:          false,
			IsClosed:          false,
		},
	}

	assert.Equal(t, expectedNewWalletResponses, responses.NewWalletsResponses)
	assert.Equal(t, decimal.NewFromFloat(62.41).String(), responses.OverallTotalUSDBalance.String())
}

// Test for zero rate: Ensures handling of wallets with zero rate currencies.
func TestWalletListZeroRate(t *testing.T) {
	t.Parallel()

	currencies := []types.Currency{
		{
			Code:      "ngn",
			RateBuy:   decimal.Zero,
			Precision: 2,
		},
	}

	// Arrange
	wallets := []*types.Wallet{
		{
			ID:               "1",
			CustomerID:       "customer1",
			CurrencyCode:     "NGN",
			AvailableBalance: decimal.NewFromFloat(1000),
			LienBalance:      decimal.NewFromFloat(0),
		},
	}

	responses, err := WalletList(wallets, currencies)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, responses.OverallTotalUSDBalance.String(), decimal.Zero.String())
	assert.NotEmpty(t, responses)
}

func TestTotalBalancesCurrencyNotFound(t *testing.T) {
	t.Parallel()

	// Arrange
	totalAmountUSD := decimal.NewFromFloat(100)
	currencies := []types.Currency{{Code: "EUR"}}

	// Act
	balances, err := TotalBalances(totalAmountUSD, currencies)

	// Assert
	assert.Error(t, err)
	assert.Empty(t, balances)
}

func TestTotalBalancesSuccess(t *testing.T) {
	// Arrange
	totalAmountUSD := decimal.NewFromFloat(100)
	currencies := []types.Currency{
		{
			Code: "NGN", RateSell: decimal.NewFromFloat(415), Precision: 2, IconURL: "https://ngn.icon", Symbol: "₦",
		},
		{
			Code: "USD", RateSell: decimal.NewFromFloat(1), Precision: 2, IconURL: "https://usd.icon", Symbol: "$",
		},
	}

	// Act
	balances, err := TotalBalances(totalAmountUSD, currencies)
	assert.NoError(t, err)

	expectedBalances := []TotalBalanceResponse{
		{
			CurrencyCode: "NGN", CurrencySymbol: "₦",
			LogoURL: "https://ngn.icon", Total: decimal.NewFromFloat(41500).StringFixedBank(2),
		},
		{
			CurrencyCode: "USD", CurrencySymbol: "$",
			LogoURL: "https://usd.icon", Total: decimal.NewFromFloat(100).StringFixedBank(2),
		},
	}

	assert.Equal(t, expectedBalances, balances)
}
