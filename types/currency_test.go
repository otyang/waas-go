package types

import (
	"errors"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestNewRateCalculator(t *testing.T) {
	baseCurrency := "USD"
	currencies := []CurrencyInfo{
		{Code: "USD", SpreadMarginBuy: decimal.NewFromFloat(0.01), SpreadMarginSell: decimal.NewFromFloat(0.01)},
		{Code: "EUR", SpreadMarginBuy: decimal.NewFromFloat(0.02), SpreadMarginSell: decimal.NewFromFloat(0.02)},
	}
	rates := map[string]float64{
		"USD": 1.0,
		"EUR": 0.85,
	}

	t.Run("successful creation", func(t *testing.T) {
		rc, err := NewRateCalculator(baseCurrency, currencies, rates)
		assert.NoError(t, err)
		assert.Equal(t, baseCurrency, rc.GetBaseCurrency())
	})

	t.Run("empty base currency", func(t *testing.T) {
		_, err := NewRateCalculator("", currencies, rates)
		assert.Equal(t, ErrBaseCurrencyNotFound, err)
	})

	t.Run("empty currencies", func(t *testing.T) {
		_, err := NewRateCalculator(baseCurrency, []CurrencyInfo{}, rates)
		assert.Equal(t, ErrEmptyCurrencySource, err)
	})

	t.Run("empty rates", func(t *testing.T) {
		_, err := NewRateCalculator(baseCurrency, currencies, map[string]float64{})
		assert.Equal(t, ErrEmptyCurrencySource, err)
	})
}

func TestCalculateExchangeRate(t *testing.T) {
	baseCurrency := "USD"
	currencies := []CurrencyInfo{
		{
			Code:             "USD",
			SpreadMarginBuy:  decimal.NewFromFloat(0.01), // 1% buy spread
			SpreadMarginSell: decimal.NewFromFloat(0.01), // 1% sell spread
			Precision:        2,
		},
		{
			Code:             "EUR",
			SpreadMarginBuy:  decimal.NewFromFloat(0.02),  // 2% buy spread
			SpreadMarginSell: decimal.NewFromFloat(0.015), // 1.5% sell spread
			Precision:        2,
		},
		{
			Code:             "GBP",
			SpreadMarginBuy:  decimal.NewFromFloat(0.015), // 1.5% buy spread
			SpreadMarginSell: decimal.NewFromFloat(0.02),  // 2% sell spread
			Precision:        2,
		},
	}
	rates := map[string]float64{
		"USD": 1.0,
		"EUR": 0.85, // 1 USD = 0.85 EUR
		"GBP": 0.75, // 1 USD = 0.75 GBP
	}

	rc, err := NewRateCalculator(baseCurrency, currencies, rates)
	assert.NoError(t, err)

	t.Run("base to foreign currency", func(t *testing.T) {
		rate, err := rc.CalculateExchangeRate("USD", "EUR")
		assert.NoError(t, err)

		// Expected mid rate: 0.85
		// Buy rate (customer buys EUR): mid * (1 + 0.02) = 0.85 * 1.02 = 0.867
		expectedBuy := decimal.NewFromFloat(0.867)
		assert.True(t, rate.BuyRate.Equal(expectedBuy), "expected %s, got %s", expectedBuy, rate.BuyRate)

		// Sell rate not applicable for base to foreign
	})

	t.Run("foreign to base currency", func(t *testing.T) {
		rate, err := rc.CalculateExchangeRate("EUR", "USD")
		assert.NoError(t, err)

		// Expected mid rate: 1/0.85 = ~1.17647
		// Sell rate (customer sells EUR): mid * (1 - 0.015) = 1.17647 * 0.985 ≈ 1.15882
		expectedSell := decimal.NewFromFloat(1.15882)
		assert.True(t, rate.SellRate.Equal(expectedSell.Round(5)), "expected %s, got %s", expectedSell, rate.SellRate)
	})

	t.Run("cross currency conversion", func(t *testing.T) {
		rate, err := rc.CalculateExchangeRate("EUR", "GBP")
		assert.NoError(t, err)

		// Expected mid rate: (1/0.85) * 0.75 ≈ 0.88235
		// Buy rate: mid * (1 + 0.015) ≈ 0.88235 * 1.015 ≈ 0.89558
		// Sell rate: mid * (1 - 0.015) ≈ 0.88235 * 0.985 ≈ 0.86911
		expectedBuy := decimal.NewFromFloat(0.89558)
		expectedSell := decimal.NewFromFloat(0.86911)
		assert.True(t, rate.BuyRate.Equal(expectedBuy.Round(5)), "expected %s, got %s", expectedBuy, rate.BuyRate)
		assert.True(t, rate.SellRate.Equal(expectedSell.Round(5)), "expected %s, got %s", expectedSell, rate.SellRate)
	})

	t.Run("same currency", func(t *testing.T) {
		_, err := rc.CalculateExchangeRate("USD", "USD")
		assert.Equal(t, ErrSameCurrency, err)
	})

	t.Run("invalid currency", func(t *testing.T) {
		_, err := rc.CalculateExchangeRate("USD", "XYZ")
		assert.ErrorContains(t, err, "currency not found: XYZ")
	})

	t.Run("missing rate", func(t *testing.T) {
		_, err := rc.CalculateExchangeRate("USD", "EUR")
		assert.NoError(t, err) // Should pass since EUR rate exists

		// Add test with missing rate
		badRc, _ := NewRateCalculator(baseCurrency, currencies, map[string]float64{"USD": 1.0})
		_, err = badRc.CalculateExchangeRate("USD", "EUR")
		assert.ErrorContains(t, err, "rate for EUR not found")
	})
}

func TestGetSimpleRate(t *testing.T) {
	baseCurrency := "USD"
	currencies := []CurrencyInfo{{Code: "USD"}, {Code: "EUR"}}
	rates := map[string]float64{
		"USD": 1.0,
		"EUR": 0.85,
	}

	rc, err := NewRateCalculator(baseCurrency, currencies, rates)
	assert.NoError(t, err)

	t.Run("existing currency", func(t *testing.T) {
		rate, err := rc.GetSimpleRate("EUR")
		assert.NoError(t, err)
		assert.True(t, rate.Equal(decimal.NewFromFloat(0.85)))
	})

	t.Run("non-existent currency", func(t *testing.T) {
		_, err := rc.GetSimpleRate("GBP")
		assert.ErrorContains(t, err, "currency not found: GBP")
	})
}

func TestNewQuote(t *testing.T) {
	currencies := []CurrencyInfo{
		{
			Code:             "USD",
			Precision:        2,
			SpreadMarginBuy:  decimal.NewFromFloat(0.01),
			SpreadMarginSell: decimal.NewFromFloat(0.01),
		},
		{
			Code:             "EUR",
			Precision:        2,
			SpreadMarginBuy:  decimal.NewFromFloat(0.02),
			SpreadMarginSell: decimal.NewFromFloat(0.015),
		},
	}
	rates := map[string]float64{
		"USD": 1.0,
		"EUR": 0.85,
	}

	t.Run("successful fixed fee quote", func(t *testing.T) {
		quote, err := NewQuote(
			currencies,
			rates,
			"USD",
			"USD",
			"EUR",
			decimal.NewFromFloat(100),
			decimal.NewFromFloat(5),
			NewQuoteFeeTypeFixed,
		)
		assert.NoError(t, err)
		assert.Equal(t, "USD", quote.FromCurrency)
		assert.Equal(t, "EUR", quote.ToCurrency)
		assert.Equal(t, decimal.NewFromFloat(100), quote.FromAmount)
		assert.Equal(t, decimal.NewFromFloat(5), quote.Fee)
		assert.True(t, quote.ToAmount.GreaterThan(decimal.Zero))
		assert.Equal(t, NewQuoteFeeTypeFixed, quote.Metadata["feeType"])
	})

	t.Run("successful percentage fee quote", func(t *testing.T) {
		quote, err := NewQuote(
			currencies,
			rates,
			"USD",
			"USD",
			"EUR",
			decimal.NewFromFloat(100),
			decimal.NewFromFloat(1), // 1%
			NewQuoteFeeTypePercentage,
		)
		assert.NoError(t, err)
		assert.Equal(t, decimal.NewFromFloat(1), quote.Fee)
		assert.Equal(t, NewQuoteFeeTypePercentage, quote.Metadata["feeType"])
	})

	t.Run("invalid fee type", func(t *testing.T) {
		_, err := NewQuote(
			currencies,
			rates,
			"USD",
			"USD",
			"EUR",
			decimal.NewFromFloat(100),
			decimal.NewFromFloat(1),
			"invalid",
		)
		assert.ErrorContains(t, err, "invalid fee type")
	})

	t.Run("negative fee", func(t *testing.T) {
		_, err := NewQuote(
			currencies,
			rates,
			"USD",
			"USD",
			"EUR",
			decimal.NewFromFloat(100),
			decimal.NewFromFloat(-1),
			NewQuoteFeeTypeFixed,
		)
		assert.ErrorContains(t, err, "fee cannot be negative")
	})

	t.Run("invalid currency pair", func(t *testing.T) {
		_, err := NewQuote(
			currencies,
			rates,
			"USD",
			"",
			"EUR",
			decimal.NewFromFloat(100),
			decimal.NewFromFloat(1),
			NewQuoteFeeTypeFixed,
		)
		assert.Equal(t, ErrInvalidCurrencyPair, err)
	})

	t.Run("currency not found", func(t *testing.T) {
		_, err := NewQuote(
			currencies,
			rates,
			"USD",
			"USD",
			"GBP", // Not in our test data
			decimal.NewFromFloat(100),
			decimal.NewFromFloat(1),
			NewQuoteFeeTypeFixed,
		)
		assert.ErrorContains(t, err, "currency not found: GBP")
	})

	t.Run("fee exceeds amount", func(t *testing.T) {
		quote, err := NewQuote(
			currencies,
			rates,
			"USD",
			"USD",
			"EUR",
			decimal.NewFromFloat(100),
			decimal.NewFromFloat(150), // Fixed fee > amount
			NewQuoteFeeTypeFixed,
		)
		assert.NoError(t, err)
		assert.True(t, quote.Fee.Equal(decimal.NewFromFloat(100)))
		assert.True(t, quote.ToAmount.Equal(decimal.Zero))
	})
}

//====

func TestFindCurrencyInfo(t *testing.T) {
	// Setup test data
	testCurrencies := []CurrencyInfo{
		{
			Code:      "USD",
			Name:      "US Dollar",
			Symbol:    "$",
			Precision: 2,
		},
		{
			Code:      "EUR",
			Name:      "Euro",
			Symbol:    "€",
			Precision: 2,
		},
		{
			Code:      "JPY",
			Name:      "Japanese Yen",
			Symbol:    "¥",
			Precision: 0,
		},
	}

	t.Run("successful find - exact match", func(t *testing.T) {
		currency, err := FindCurrencyInfo(testCurrencies, "USD")
		assert.NoError(t, err)
		assert.Equal(t, "USD", currency.Code)
		assert.Equal(t, "US Dollar", currency.Name)
	})

	t.Run("successful find - case insensitive", func(t *testing.T) {
		currency, err := FindCurrencyInfo(testCurrencies, "eur")
		assert.NoError(t, err)
		assert.Equal(t, "EUR", currency.Code)
		assert.Equal(t, "Euro", currency.Name)
	})

	t.Run("successful find - with whitespace", func(t *testing.T) {
		currency, err := FindCurrencyInfo(testCurrencies, "  jpy  ")
		assert.NoError(t, err)
		assert.Equal(t, "JPY", currency.Code)
		assert.Equal(t, "Japanese Yen", currency.Name)
	})

	t.Run("currency not found", func(t *testing.T) {
		_, err := FindCurrencyInfo(testCurrencies, "GBP")
		assert.Error(t, err)
		assert.Equal(t, "currency not found: GBP", err.Error())
	})

	t.Run("empty currency code", func(t *testing.T) {
		_, err := FindCurrencyInfo(testCurrencies, "")
		assert.Error(t, err)
		assert.Equal(t, ErrCurrencyNotFound, errors.Unwrap(err))
	})

	t.Run("empty currencies list", func(t *testing.T) {
		_, err := FindCurrencyInfo([]CurrencyInfo{}, "USD")
		assert.Error(t, err)
		assert.Equal(t, "currency not found: USD", err.Error())
	})

	t.Run("verify copy not reference", func(t *testing.T) {
		// Get the original and found currency
		originalIndex := 1 // EUR
		original := testCurrencies[originalIndex]
		found, err := FindCurrencyInfo(testCurrencies, "EUR")
		assert.NoError(t, err)

		// Modify the found currency
		found.Name = "Modified Euro"

		// Verify original wasn't changed
		assert.Equal(t, "Euro", testCurrencies[originalIndex].Name)
		assert.Equal(t, "Modified Euro", found.Name)
		assert.NotSame(t, &original, found)
	})
}
