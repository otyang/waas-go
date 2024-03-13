package currency

import (
	"fmt"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestFindCurrency(t *testing.T) {
	t.Parallel()

	// Setup Currencies
	currencies := []Currency{
		{Code: "USD", RateBuy: decimal.NewFromInt(100), RateSell: decimal.NewFromInt(101)},
		{Code: "EUR", RateBuy: decimal.NewFromInt(120), RateSell: decimal.NewFromInt(121)},
	}

	// Valid code
	currency, err := FindCurrency(currencies, "USD")
	assert.NoError(t, err)
	assert.Equal(t, currency.Code, "USD")

	// Invalid code
	_, err = FindCurrency(currencies, "GBP")
	assert.Error(t, err)
	assert.Equal(t, fmt.Errorf(ErrCurrencyNotFound, "GBP"), err)

	// Empty currency source
	_, err = FindCurrency(nil, "USD")
	assert.Error(t, err)
	assert.Equal(t, ErrEmptyCurrencySource, err)
}

func TestCalculateRate(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		base      string
		from      string
		to        string
		expected  decimal.Decimal
		shouldErr bool
	}{
		{
			name:     "Same currency conversion",
			base:     "USD",
			from:     "USD",
			to:       "USD",
			expected: decimal.NewFromInt(1),
		},
		{
			name:     "Base currency to target currency conversion",
			base:     "USD",
			from:     "USD",
			to:       "EUR",
			expected: decimal.NewFromFloat(1.18),
		},
		{
			name:     "Target currency to base currency conversion",
			base:     "USD",
			from:     "EUR",
			to:       "USD",
			expected: decimal.NewFromFloat(1.18),
		},
		{
			name:     "Cross-rate conversion",
			base:     "USD",
			from:     "EUR",
			to:       "JPY",
			expected: decimal.NewFromFloat(145.24),
		},
		{
			name:      "Non-existent base currency",
			base:      "ZZZ",
			from:      "USD",
			to:        "EUR",
			shouldErr: true,
		},
		{
			name:      "Non-existent from currency",
			base:      "USD",
			from:      "ZZZ",
			to:        "EUR",
			shouldErr: true,
		},
		{
			name:      "Non-existent to currency",
			base:      "USD",
			from:      "EUR",
			to:        "ZZZ",
			shouldErr: true,
		},
	}

	// Test currencies
	currencies := []Currency{
		{Code: "USD", Precision: 2, RateBuy: decimal.NewFromFloat(1.0), RateSell: decimal.NewFromFloat(1.0)},
		{Code: "EUR", Precision: 2, RateBuy: decimal.NewFromFloat(0.85), RateSell: decimal.NewFromFloat(1.18)},
		{Code: "JPY", Precision: 2, RateBuy: decimal.NewFromFloat(0.0081), RateSell: decimal.NewFromFloat(123.45)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualRate, err := CalculateRate(currencies, tc.base, tc.from, tc.to)

			if tc.shouldErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tc.expected.String(), actualRate.RoundCeil(2).String())
			assert.True(t, tc.expected.Equal(actualRate.RoundCeil(2)))
		})
	}
}

func TestNewQuote_ValidInput(t *testing.T) {
	t.Parallel()

	// Set up test data
	var (
		rateSource = []Currency{
			{Code: "USD", Precision: 2, RateBuy: decimal.NewFromInt(1), RateSell: decimal.NewFromInt(1)},
			{Code: "EUR", Precision: 2, RateBuy: decimal.NewFromFloat(0.95), RateSell: decimal.NewFromFloat(0.95)},
		}
		baseCurrency = "USD"
		fromCurrency = "USD"
		toCurrency   = "EUR"
		fromAmount   = decimal.NewFromFloat(100)
		fee          = decimal.NewFromFloat(5)
	)

	rate, err := CalculateRate(rateSource, baseCurrency, fromCurrency, toCurrency)
	assert.NoError(t, err)

	want := &Quote{
		BaseCurrency: baseCurrency,
		FromCurrency: fromCurrency,
		FromAmount:   fromAmount,
		Fee:          fee,
		GrossAmount:  fromAmount.Add(fee),
		Rate:         rate,
		ToCurrency:   toCurrency,
		ToAmount:     decimal.NewFromFloat(95),
		Date:         time.Time{},
	}

	// Call the function
	got, err := NewQuote(rateSource, baseCurrency, fromCurrency, toCurrency, fromAmount, fee)
	assert.NoError(t, err)
	assert.Equal(t, want.BaseCurrency, got.BaseCurrency)
	assert.Equal(t, want.FromCurrency, got.FromCurrency)
	assert.Equal(t, want.FromAmount, got.FromAmount)
	assert.Equal(t, want.Fee, got.Fee)
	assert.Equal(t, want.GrossAmount, got.GrossAmount)
	assert.Equal(t, want.Rate, got.Rate)
	assert.Equal(t, want.ToCurrency, got.ToCurrency)
	assert.Equal(t, want.ToAmount.String(), got.ToAmount.String())
}

func TestNewQuote_EmptyRateSource(t *testing.T) {
	t.Parallel()

	quote, err := NewQuote(nil, "USD", "USD", "EUR", decimal.NewFromInt(100), decimal.NewFromInt(5))
	assert.Error(t, err)
	assert.Nil(t, quote)
}
