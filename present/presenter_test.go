package present

import (
	"fmt"
	"testing"
	"time"

	"github.com/otyang/waas-go"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func Test_calcTotalBalance(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                  string
		totalAmountUSD        decimal.Decimal
		currencies            []waas.Currency
		expectedTotalBalances []TotalBalanceResponse
		expectedError         error
	}{
		{
			name:           "Valid calculation",
			totalAmountUSD: decimal.NewFromInt(100),
			currencies: []waas.Currency{
				{Code: "NGN", RateSell: decimal.NewFromInt(450), Symbol: "₦", IconURL: "ngn_icon.png", Precision: 2},
				{Code: "USD", RateSell: decimal.NewFromInt(1), Symbol: "$", IconURL: "usd_icon.png", Precision: 2},
			},
			expectedTotalBalances: []TotalBalanceResponse{
				{CurrencyCode: "NGN", CurrencySymbol: "₦", LogoURL: "ngn_icon.png", Total: decimal.NewFromInt(45000)},
				{CurrencyCode: "USD", CurrencySymbol: "$", LogoURL: "usd_icon.png", Total: decimal.NewFromInt(100)},
			},
			expectedError: nil,
		},
		{
			name:           "Error finding NGN",
			totalAmountUSD: decimal.NewFromInt(50),
			currencies:     []waas.Currency{
				// ... (NGN currency missing)
			},
			expectedTotalBalances: nil,
			expectedError:         fmt.Errorf("currency not found: NGN"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualTotalBalances, err := calcTotalBalance(tc.totalAmountUSD, tc.currencies)

			if tc.expectedError == nil {
				assert.NoError(t, err)
				assert.Equal(t, actualTotalBalances, tc.expectedTotalBalances)
			}

			if tc.expectedError != nil {
				assert.Error(t, err)
			}
		})
	}
}

func Test_generateWalletResponse_NormalConversion(t *testing.T) {
	t.Parallel()

	wallet := &waas.Wallet{
		ID:               "wallet123",
		CustomerID:       "customer456",
		AvailableBalance: decimal.NewFromInt(1000),
		LienBalance:      decimal.NewFromInt(100),
		Status:           waas.WalletStatusActive,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	curr := waas.Currency{
		Code:      "BTC",
		Precision: 8,
		RateBuy:   decimal.NewFromInt(20000), // $20,000 per BTC
		RateSell:  decimal.NewFromInt(19500), // $19,500 per BTC
	}

	expectedTotalBalanceInUSD := wallet.TotalBalance().Div(curr.RateBuy).RoundCeil(int32(curr.Precision))

	response := generateWalletResponse(wallet, curr)

	// Assertions
	assert.Equal(t, response.ID, wallet.ID)
	assert.Equal(t, response.Status, wallet.Status)
	assert.Equal(t, response.CustomerID, wallet.CustomerID)
	assert.Equal(t, response.AvailableBalance.String(), decimal.NewFromInt(1000).String())
	assert.Equal(t, response.LienBalance.String(), decimal.NewFromInt(100).String())
	assert.Equal(t, response.TotalBalanceInUSD.String(), expectedTotalBalanceInUSD.String())
}
