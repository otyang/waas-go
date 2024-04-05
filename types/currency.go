package types

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

var (
	ErrCurrencyNotFound     = "currency %s not found"
	ErrEmptyCurrencySource  = errors.New("empty currency source: no rates or currency")
	ErrBaseCurrencyNotFound = errors.New("base currency not found")
)

type Currency struct {
	Code          string          `json:"code" bun:",pk"`
	Name          string          `json:"name"`
	Symbol        string          `json:"symbol"`
	IsFiat        bool            `json:"isFiat"`
	IsStableCoin  bool            `json:"isStableCoin"`
	IconURL       string          `json:"iconUrl"`
	Precision     int             `json:"precision"`
	Disabled      bool            `json:"disabled"`
	CanSell       bool            `json:"canSell"`
	CanBuy        bool            `json:"canBuy"`
	CanSwap       bool            `json:"canSwap"`
	CanDeposit    bool            `json:"canDeposit"`
	CanWithdraw   bool            `json:"canWithdraw"`
	FeeDeposit    decimal.Decimal `json:"depositfee"`
	FeeWithdrawal decimal.Decimal `json:"withdrawalfee"`
	RateBuy       decimal.Decimal `json:"rateBuy"`
	RateSell      decimal.Decimal `json:"rateSell"`
	CreatedAt     time.Time       `json:"createdAt"`
	UpdatedAt     time.Time       `json:"updatedAt"`
}

func (c *Currency) UpdateBuyRate(newRate decimal.Decimal) *Currency {
	c.RateBuy = newRate
	return c
}

func (c *Currency) UpdateSellRate(newRate decimal.Decimal) *Currency {
	c.RateSell = newRate
	return c
}

// FindCurrency finds a currency by its ISO code.
func FindCurrency(currencies []Currency, code string) (*Currency, error) {
	if len(currencies) == 0 {
		return nil, ErrEmptyCurrencySource
	}

	for i := range currencies {
		if strings.EqualFold(currencies[i].Code, code) {
			return &currencies[i], nil
		}
	}

	return nil, fmt.Errorf(ErrCurrencyNotFound, code)
}

// CalculateRate calculates the exchange rate between two currencies.
// Same Currency Conversion:
//   - from 	(you have) 			= base currency
//   - to 		(you want) 			= base currency
//   - Rate: 	1
//
// Base to Target Conversion:
//   - from 	(you have) 			= base currency
//   - to 		(you want/target) 	= another currency
//   - Rate: 	sell-rate of to
//
// Target to Base:
//   - from 	(you have/target) 	= another currency
//   - to 		(you want) 			= base currency
//   - Rate: 	1 / sell-rate of from
//
// Cross rate conversion: [Target to base and then to target]
//   - from 	(you have/source) 	= another currency
//   - to 		(you want/target) 	= another currency
//   - Rate: 	[Target to Base of: from] * [Base to Target of: to]
func calculateRate(currencies []Currency, baseCurrency, from, to string) (decimal.Decimal, error) {
	baseCurrency = strings.ToUpper(baseCurrency)
	from = strings.ToUpper(from)
	to = strings.ToUpper(to)

	// Same Currency Conversion
	if from == to {
		return decimal.NewFromInt(1), nil
	}

	_, err := FindCurrency(currencies, baseCurrency)
	if err != nil {
		return decimal.Zero, ErrBaseCurrencyNotFound
	}

	// Base to Target Currency (Sell Rate)
	if from == baseCurrency {
		toCurrency, err := FindCurrency(currencies, to)
		if err != nil {
			return decimal.Zero, err
		}

		return toCurrency.RateSell, nil
	}

	// Target to Base Currency (Buy Rate)
	if to == baseCurrency {
		fromCurrency, err := FindCurrency(currencies, from)
		if err != nil {
			return decimal.Zero, err
		}

		if fromCurrency.RateBuy.Equal(decimal.Zero) { // since mathematically 1 divide by 0 is error
			return decimal.Zero, nil
		}
		return decimal.NewFromInt(1).Div(fromCurrency.RateBuy), nil
	}

	// Cross Rate Conversion
	fromCurrency, err := FindCurrency(currencies, from)
	if err != nil {
		return decimal.Zero, err
	}
	toCurrency, err := FindCurrency(currencies, to)
	if err != nil {
		return decimal.Zero, err
	}

	// (target to base) to target
	if fromCurrency.RateBuy.Equal(decimal.Zero) { // since 1 divide by 0 is error. lets avoid it
		return decimal.Zero, nil
	}
	return (decimal.NewFromInt(1).Div(fromCurrency.RateBuy)).Mul(toCurrency.RateSell), nil
}

// Quote structure
type Quote struct {
	BaseCurrency string          `json:"baseCurrency"`
	FromCurrency string          `json:"fromCurrency"`
	FromAmount   decimal.Decimal `json:"fromAmount"`
	ToCurrency   string          `json:"toCurrency"`
	ToAmount     decimal.Decimal `json:"toAmount"`
	Fee          decimal.Decimal `json:"fee"`
	Rate         decimal.Decimal `json:"rate"`
	GrossAmount  decimal.Decimal `json:"grossAmount"`
	Date         time.Time       `json:"date"`
}

// NewQuote creates a new quote object.
func NewQuote(rateSource []Currency, baseCurrency, fromCurrency, toCurrency string, fromAmount, fee decimal.Decimal) (*Quote, error) {
	if rateSource == nil {
		return nil, errors.New("currency object empty. shouldnt be")
	}

	rate, err := calculateRate(rateSource, baseCurrency, fromCurrency, toCurrency)
	if err != nil {
		return nil, err
	}

	infoFrom, err := FindCurrency(rateSource, fromCurrency)
	if err != nil {
		return nil, err
	}

	infoTo, err := FindCurrency(rateSource, toCurrency)
	if err != nil {
		return nil, err
	}

	return &Quote{
		BaseCurrency: baseCurrency,
		FromCurrency: fromCurrency,
		FromAmount:   fromAmount,
		ToCurrency:   toCurrency,
		ToAmount:     fromAmount.Mul(rate).RoundCeil(int32(infoTo.Precision)),
		Fee:          fee,
		Rate:         rate,
		GrossAmount:  fromAmount.Add(fee).RoundCeil(int32(infoFrom.Precision)),
		Date:         time.Now(),
	}, nil
}
