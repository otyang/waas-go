package types

import (
	"errors"
	"fmt"
	"time"

	gonanoid "github.com/matoous/go-nanoid"
	"github.com/shopspring/decimal"
)

func NewTransactionID() string {
	return GenerateID("txn_"+time.Now().UTC().Format("20060102")+"_", 8)
}

func GenerateID(prefix string, size int) string {
	return prefix + gonanoid.MustGenerate("0123456789abcdefghijklmnopqrstuvwxyz", size)
}

type WalletSummary struct {
	*Wallet       `json:"wallet"`       // Embedded Wallet
	*CurrencyInfo `json:"currencyInfo"` // Embedded CurrencyInfo
	// Computed field (not stored, calculated when needed)
	TotalBalanceInUSD decimal.Decimal `json:"totalBalanceInUSD"`
}

// GenerateWalletSummaries creates WalletSummary objects by combining wallet data with currency information
// and converting balances to USD using provided exchange rates.
//
// Parameters:
//   - wallets: List of Wallet pointers to process
//   - currencies: List of CurrencyInfo pointers for currency details
//   - exchangeRates: Map of currency codes to their USD exchange rates
//
// Returns:
//   - []*WalletSummary: List of created wallet summaries
//   - error: Aggregate error if any issues occurred (partial results may still be returned)
func GenerateWalletSummaries(
	wallets []*Wallet,
	currencies []*CurrencyInfo,
	exchangeRates map[string]float64,
) ([]*WalletSummary, error) {
	// Convert float64 rates to decimal.Decimal for precise calculations
	decimalRates := make(map[string]decimal.Decimal, len(exchangeRates))
	for code, rate := range exchangeRates {
		if rate <= 0 {
			return nil, fmt.Errorf("invalid exchange rate for %s: must be positive", code)
		}
		decimalRates[code] = decimal.NewFromFloat(rate)
	}

	// Create currency lookup map for O(1) access
	currencyMap := make(map[string]*CurrencyInfo)
	for _, ci := range currencies {
		currencyMap[ci.Code] = ci
	}

	var summaries []*WalletSummary
	var errs []error

	for _, wallet := range wallets {
		// Skip nil wallets
		if wallet == nil {
			errs = append(errs, fmt.Errorf("nil wallet encountered"))
			continue
		}

		// Find matching currency info
		currency, exists := currencyMap[wallet.CurrencyCode]
		if !exists {
			errs = append(errs, fmt.Errorf("currency info not found for wallet %s: %s", wallet.ID, wallet.CurrencyCode))
			continue
		}

		// Get exchange rate (prefer provided rates, fallback to currency's rate)
		rate, rateExists := decimalRates[wallet.CurrencyCode]
		if !rateExists {
			if currency.IsFiat && currency.AutomaticUpdate {
				errs = append(errs, fmt.Errorf("exchange rate required for fiat currency %s in wallet %s",
					wallet.CurrencyCode, wallet.ID))
				continue
			}
			rate = decimal.NewFromFloat(1.0) // Default for crypto/stablecoins without rate
		}

		// Create wallet summary
		summary := &WalletSummary{
			Wallet:       wallet,
			CurrencyInfo: currency,
		}

		// Calculate USD balance
		totalBalance := wallet.TotalBalance()
		if wallet.CurrencyCode == "USD" {
			summary.TotalBalanceInUSD = totalBalance
		} else {
			summary.TotalBalanceInUSD = totalBalance.Mul(rate)
		}

		summaries = append(summaries, summary)
	}

	// Return combined error if any errors occurred
	if len(errs) > 0 {
		return summaries, fmt.Errorf("%d errors occurred during processing: %v",
			len(errs), errors.Join(errs...))
	}

	return summaries, nil
}

type TotalBalanceInSpecificCurrency struct {
	Total        decimal.Decimal
	CurrencyInfo *CurrencyInfo
}

// CalculateTotalBalanceInCurrency aggregates balances from multiple WalletSummary objects
// and converts them to the specified target currency using provided exchange rates.
//
// Parameters:
//   - summaries: Slice of WalletSummary pointers to process
//   - exchangeRates: Map of currency codes to their exchange rates (relative to USD)
//   - targetCurrency: The currency code to convert all balances to (e.g., "EUR")
//   - currencies: List of available currency info for validation
//
// Returns:
//   - TotalBalanceInSpecificCurrency containing aggregated balance and currency info
//   - error if any conversion fails or currencies are invalid
func CalculateTotalBalanceInCurrency(
	summaries []*WalletSummary,
	exchangeRates map[string]float64,
	targetCurrency string,
	currencies []CurrencyInfo,
) (*TotalBalanceInSpecificCurrency, error) {
	// Validate target currency exists
	targetCurrencyInfo, err := FindCurrencyInfo(currencies, targetCurrency)
	if err != nil {
		return nil, fmt.Errorf("target currency not found: %w", err)
	}

	// Convert all rates to decimal upfront
	decimalRates := make(map[string]decimal.Decimal)
	for code, rate := range exchangeRates {
		if rate <= 0 {
			return nil, fmt.Errorf("invalid exchange rate for %s: must be positive", code)
		}
		decimalRates[code] = decimal.NewFromFloat(rate)
	}

	total := decimal.Zero
	var errs []error

	for _, summary := range summaries {
		if summary == nil {
			continue // skip nil summaries
		}

		// Get wallet balance in USD (already calculated in WalletSummary)
		balanceUSD := summary.TotalBalanceInUSD

		// Convert from USD to target currency
		if targetCurrency == "USD" {
			total = total.Add(balanceUSD)
			continue
		}

		// Get USD to target currency rate
		rate, exists := decimalRates[targetCurrency]
		if !exists {
			errs = append(errs, fmt.Errorf("exchange rate not available for target currency %s", targetCurrency))
			continue
		}

		// USD -> Target: divide by rate (since rates are USD-based)
		convertedBalance := balanceUSD.Div(rate)
		total = total.Add(convertedBalance)
	}

	if len(errs) > 0 {
		return nil, fmt.Errorf("%d conversion errors occurred: %w", len(errs), errors.Join(errs...))
	}

	return &TotalBalanceInSpecificCurrency{
		Total:        total,
		CurrencyInfo: targetCurrencyInfo,
	}, nil
}
