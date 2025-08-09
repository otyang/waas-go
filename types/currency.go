package types

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

// Error definitions for currency operations
var (
	ErrCurrencyNotFound     = errors.New("currency not found")
	ErrEmptyCurrencySource  = errors.New("empty currency source")
	ErrInvalidCurrencyPair  = errors.New("invalid currency pair")
	ErrRateCalculation      = errors.New("rate calculation error")
	ErrSameCurrency         = errors.New("cannot convert between same currency")
	ErrBaseCurrencyNotFound = errors.New("base currency not found")
)

// CurrencyInfo represents a financial currency with all its properties
type CurrencyInfo struct {
	Code             string          `json:"code" bun:",pk"`   // ISO currency code (e.g., "USD")
	Name             string          `json:"name"`             // Full currency name
	Symbol           string          `json:"symbol"`           // Currency symbol (e.g., "$")
	IsFiat           bool            `json:"isFiat"`           // Whether it's a fiat currency
	IsStableCoin     bool            `json:"isStableCoin"`     // Whether it's a stablecoin
	IconURL          string          `json:"iconUrl"`          // URL to currency icon
	Precision        int             `json:"precision"`        // Decimal precision for calculations
	Disabled         bool            `json:"disabled"`         // Whether currency is disabled
	CanSell          bool            `json:"canSell"`          // Whether selling is allowed
	CanBuy           bool            `json:"canBuy"`           // Whether buying is allowed
	CanSwap          bool            `json:"canSwap"`          // Whether swapping is allowed
	CanDeposit       bool            `json:"canDeposit"`       // Whether deposits are allowed
	CanWithdraw      bool            `json:"canWithdraw"`      // Whether withdrawals are allowed
	FeeDeposit       decimal.Decimal `json:"depositFee"`       // Deposit fee amount
	FeeWithdrawal    decimal.Decimal `json:"withdrawalFee"`    // Withdrawal fee amount
	SpreadMarginBuy  decimal.Decimal `json:"spreadMarginBuy"`  // Buy spread margin
	SpreadMarginSell decimal.Decimal `json:"spreadMarginSell"` // Sell spread margin
	AutomaticUpdate  bool            `json:"automaticUpdate"`  // Whether rates update automatically
	CreatedAt        time.Time       `json:"createdAt"`        // When currency was added
	UpdatedAt        time.Time       `json:"updatedAt"`        // Last update timestamp
}

// FindCurrencyInfo searches for a currency in the given list by its code (case-insensitive)
// Parameters:
//   - currencies: List of CurrencyInfo to search through
//   - currencyCode: The currency code to find (e.g., "USD", "EUR")
//
// Returns:
//   - *CurrencyInfo if found
//   - error if not found
func FindCurrencyInfo(currencies []CurrencyInfo, currencyCode string) (*CurrencyInfo, error) {
	currencyCode = strings.ToUpper(strings.TrimSpace(currencyCode))
	if currencyCode == "" {
		return nil, ErrCurrencyNotFound
	}

	for _, currency := range currencies {
		if strings.EqualFold(currency.Code, currencyCode) {
			// Return a copy to avoid modifying the original
			found := currency
			return &found, nil
		}
	}

	return nil, fmt.Errorf("%w: %s", ErrCurrencyNotFound, currencyCode)
}

// ExchangeRate represents a currency exchange rate with buy/sell prices
type ExchangeRate struct {
	CurrencyPair string          `json:"currencyPair"` // Pair in "FROM/TO" format
	FromCurrency string          `json:"fromCurrency"` // Source currency code
	ToCurrency   string          `json:"toCurrency"`   // Target currency code
	RateType     string          `json:"rateType"`     // Rate type (e.g., "spot")
	BuyRate      decimal.Decimal `json:"buyRate"`      // Rate for buying the target currency
	SellRate     decimal.Decimal `json:"sellRate"`     // Rate for selling the source currency
	UpdatedAt    time.Time       `json:"updatedAt"`    // When rate was last updated
	Source       string          `json:"source"`       // Rate source (e.g., "ECB")
}

// Quote represents a currency conversion quote
type Quote struct {
	BaseCurrency     string            `json:"baseCurrency"`        // System's base currency
	FromCurrency     string            `json:"fromCurrency"`        // Currency to convert from
	FromAmount       decimal.Decimal   `json:"fromAmount"`          // Original amount to convert
	ToCurrency       string            `json:"toCurrency"`          // Currency to convert to
	ToAmount         decimal.Decimal   `json:"toAmount"`            // Converted amount
	NetAmount        decimal.Decimal   `json:"netAmount"`           // Amount after fees
	Fee              decimal.Decimal   `json:"fee"`                 // Applied fee amount
	Rate             decimal.Decimal   `json:"rate"`                // Exchange rate used
	Date             time.Time         `json:"date"`                // Quote generation time
	FromCurrencyInfo CurrencyInfo      `json:"fromCurrencyInfo"`    // Source currency details
	ToCurrencyInfo   CurrencyInfo      `json:"toCurrencyInfo"`      // Target currency details
	Metadata         map[string]string `json:"metadata,omitempty"`  // Additional data
	ExpiresAt        *time.Time        `json:"expiresAt,omitempty"` // Quote expiration
	QuoteID          string            `json:"quoteId,omitempty"`   // Unique identifier
	RateType         string            `json:"rateType,omitempty"`  // Rate type used
}

// RateCalculator handles currency rate calculations with spreads and margins
type RateCalculator struct {
	baseCurrency string             // System's base currency code (e.g., "USD")
	currencies   []CurrencyInfo     // List of supported currencies
	rates        map[string]float64 // Current exchange rates
}

// NewRateCalculator creates a new RateCalculator instance
// Parameters:
//   - baseCurrency: The system's base currency code
//   - currencies: List of supported currencies
//   - rates: Current exchange rates map (currency code -> rate)
//
// Returns:
//   - *RateCalculator instance
//   - error if validation fails
func NewRateCalculator(baseCurrency string, currencies []CurrencyInfo, rates map[string]float64) (*RateCalculator, error) {
	if baseCurrency == "" {
		return nil, ErrBaseCurrencyNotFound
	}
	if len(currencies) == 0 {
		return nil, ErrEmptyCurrencySource
	}
	if len(rates) == 0 {
		return nil, ErrEmptyCurrencySource
	}

	return &RateCalculator{
		baseCurrency: strings.ToUpper(baseCurrency),
		currencies:   currencies,
		rates:        rates,
	}, nil
}

// CalculateExchangeRate calculates the exchange rate between two currencies including spreads
// Parameters:
//   - fromCurrency: Source currency code
//   - toCurrency: Target currency code
//
// Returns:
//   - *ExchangeRate containing buy/sell rates
//   - error if calculation fails
func (rc *RateCalculator) CalculateExchangeRate(fromCurrency, toCurrency string) (*ExchangeRate, error) {
	// Normalize and validate input
	fromCurrency = strings.ToUpper(strings.TrimSpace(fromCurrency))
	toCurrency = strings.ToUpper(strings.TrimSpace(toCurrency))

	if fromCurrency == "" || toCurrency == "" {
		return nil, ErrInvalidCurrencyPair
	}
	if fromCurrency == toCurrency {
		return nil, ErrSameCurrency
	}

	// Get currency information
	fromInfo, err := rc.findCurrencyInfo(fromCurrency)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrCurrencyNotFound, fromCurrency)
	}
	toInfo, err := rc.findCurrencyInfo(toCurrency)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrCurrencyNotFound, toCurrency)
	}

	// Get current rates
	fromRate, ok := rc.rates[fromCurrency]
	if !ok {
		return nil, fmt.Errorf("%w: rate for %s not found", ErrCurrencyNotFound, fromCurrency)
	}
	toRate, ok := rc.rates[toCurrency]
	if !ok {
		return nil, fmt.Errorf("%w: rate for %s not found", ErrCurrencyNotFound, toCurrency)
	}

	// Convert to decimal for precise calculations
	fromRateDec := decimal.NewFromFloat(fromRate)
	toRateDec := decimal.NewFromFloat(toRate)

	// Calculate mid market rate (without spreads)
	var midRate decimal.Decimal
	switch {
	case fromCurrency == rc.baseCurrency:
		// Direct conversion from base currency
		midRate = toRateDec
	case toCurrency == rc.baseCurrency:
		// Inverse conversion to base currency
		midRate = decimal.NewFromInt(1).Div(fromRateDec)
	default:
		// Cross-currency conversion (neither is base)
		midRate = decimal.NewFromInt(1).Div(fromRateDec).Mul(toRateDec)
	}

	// Apply spreads to get buy/sell rates
	buyRate := midRate.Mul(decimal.NewFromInt(1).Add(toInfo.SpreadMarginBuy))
	sellRate := midRate.Mul(decimal.NewFromInt(1).Sub(fromInfo.SpreadMarginSell))

	return &ExchangeRate{
		CurrencyPair: fmt.Sprintf("%s/%s", fromCurrency, toCurrency),
		FromCurrency: fromCurrency,
		ToCurrency:   toCurrency,
		BuyRate:      buyRate,
		SellRate:     sellRate,
		UpdatedAt:    time.Now(),
	}, nil
}

// findCurrencyInfo finds currency info by code (case-insensitive)
func (rc *RateCalculator) findCurrencyInfo(currencyCode string) (*CurrencyInfo, error) {
	for _, currency := range rc.currencies {
		if strings.EqualFold(currency.Code, currencyCode) {
			return &currency, nil
		}
	}
	return nil, ErrCurrencyNotFound
}

// GetSimpleRate gets the basic rate without spreads
func (rc *RateCalculator) GetSimpleRate(currencyCode string) (decimal.Decimal, error) {
	rate, ok := rc.rates[strings.ToUpper(currencyCode)]
	if !ok {
		return decimal.Zero, fmt.Errorf("%w: %s", ErrCurrencyNotFound, currencyCode)
	}
	return decimal.NewFromFloat(rate), nil
}

// GetBaseCurrency returns the calculator's base currency
func (rc *RateCalculator) GetBaseCurrency() string {
	return rc.baseCurrency
}

// FeeType constants
const (
	NewQuoteFeeTypeFixed      = "fixed"
	NewQuoteFeeTypePercentage = "percentage"
)

// NewQuote creates a new currency conversion quote with fee type support
// Parameters:
//   - ci: List of available currencies
//   - rates: Current exchange rates
//   - baseCurrency: System's base currency
//   - fromCurrency: Source currency
//   - toCurrency: Target currency
//   - fromAmount: Amount to convert
//   - fee: Conversion fee (either fixed amount or percentage)
//   - feeType: Type of fee ("fixed" or "percentage")
//
// Returns:
//   - *Quote containing conversion details
//   - error if conversion fails
func NewQuote(
	ci []CurrencyInfo,
	rates map[string]float64,
	baseCurrency, fromCurrency, toCurrency string,
	fromAmount, fee decimal.Decimal,
	feeType string,
) (*Quote, error) {
	// Validate inputs
	if len(ci) == 0 {
		return nil, ErrEmptyCurrencySource
	}
	if len(rates) == 0 {
		return nil, ErrEmptyCurrencySource
	}
	if baseCurrency == "" {
		return nil, ErrBaseCurrencyNotFound
	}
	if fromCurrency == "" || toCurrency == "" {
		return nil, ErrInvalidCurrencyPair
	}
	if fromAmount.LessThanOrEqual(decimal.Zero) {
		return nil, errors.New("from amount must be positive")
	}
	if fee.LessThan(decimal.Zero) {
		return nil, errors.New("fee cannot be negative")
	}
	if feeType != NewQuoteFeeTypeFixed && feeType != NewQuoteFeeTypePercentage {
		return nil, errors.New("invalid fee type, must be 'fixed' or 'percentage'")
	}

	// Normalize currency codes
	baseCurrency = strings.ToUpper(baseCurrency)
	fromCurrency = strings.ToUpper(fromCurrency)
	toCurrency = strings.ToUpper(toCurrency)

	// Find currency information
	var fromInfo, toInfo *CurrencyInfo
	for _, currency := range ci {
		if strings.EqualFold(currency.Code, fromCurrency) {
			fromInfo = &currency
		}
		if strings.EqualFold(currency.Code, toCurrency) {
			toInfo = &currency
		}
	}

	if fromInfo == nil {
		return nil, fmt.Errorf("%w: %s", ErrCurrencyNotFound, fromCurrency)
	}
	if toInfo == nil {
		return nil, fmt.Errorf("%w: %s", ErrCurrencyNotFound, toCurrency)
	}

	// Calculate actual fee amount based on fee type
	var actualFee decimal.Decimal
	switch feeType {
	case NewQuoteFeeTypePercentage:
		if fee.GreaterThan(decimal.NewFromInt(100)) {
			return nil, errors.New("percentage fee cannot exceed 100%")
		}
		actualFee = fromAmount.Mul(fee.Div(decimal.NewFromInt(100)))
	case NewQuoteFeeTypeFixed:
		actualFee = fee
	}

	// Ensure fee doesn't exceed the fromAmount
	if actualFee.GreaterThan(fromAmount) {
		actualFee = fromAmount
	}

	// Create rate calculator
	calculator, err := NewRateCalculator(baseCurrency, ci, rates)
	if err != nil {
		return nil, fmt.Errorf("failed to create rate calculator: %w", err)
	}

	// Calculate exchange rate
	exchangeRate, err := calculator.CalculateExchangeRate(fromCurrency, toCurrency)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate exchange rate: %w", err)
	}

	// Determine appropriate rate based on conversion direction
	var rate decimal.Decimal
	switch {
	case fromCurrency == baseCurrency:
		rate = exchangeRate.BuyRate // Buying target currency
	case toCurrency == baseCurrency:
		rate = exchangeRate.SellRate // Selling source currency
	default:
		rate = exchangeRate.BuyRate.Add(exchangeRate.SellRate).Div(decimal.NewFromInt(2)) // Cross-currency
	}

	// Calculate amount after fee deduction
	amountAfterFee := fromAmount.Sub(actualFee)
	if amountAfterFee.LessThan(decimal.Zero) {
		amountAfterFee = decimal.Zero
	}

	// Calculate final converted amount with proper rounding
	toAmount := amountAfterFee.Mul(rate).Round(int32(toInfo.Precision))

	// Build and return quote
	return &Quote{
		BaseCurrency:     baseCurrency,
		FromCurrency:     fromCurrency,
		FromAmount:       fromAmount,
		ToCurrency:       toCurrency,
		ToAmount:         toAmount,
		NetAmount:        toAmount,
		Fee:              actualFee,
		Rate:             rate,
		Date:             time.Now(),
		FromCurrencyInfo: *fromInfo,
		ToCurrencyInfo:   *toInfo,
		Metadata: map[string]string{
			"feeType": feeType,
		},
	}, nil
}
