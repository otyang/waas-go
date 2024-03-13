package account

import (
	"context"
	"testing"
	"time"

	"github.com/otyang/waas-go/currency"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestAccount_CreateCurrency(t *testing.T) {
	t.Parallel()

	acc := &Account{
		db: TestDB,
	}

	gotNGN, err := acc.CreateCurrency(context.Background(), currency.Currency{
		Code:          "NGN",
		Name:          "",
		Symbol:        "",
		IsFiat:        false,
		IsStableCoin:  false,
		IconURL:       "",
		Precision:     0,
		Disabled:      false,
		CanSell:       false,
		CanBuy:        false,
		CanSwap:       false,
		CanDeposit:    false,
		CanWithdraw:   false,
		FeeDeposit:    decimal.Decimal{},
		FeeWithdrawal: decimal.Decimal{},
		RateBuy:       decimal.Decimal{},
		RateSell:      decimal.Decimal{},
		CreatedAt:     time.Time{},
		UpdatedAt:     time.Time{},
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, gotNGN)

	// update
	gotNGN.Name = "Nigerian Naira"
	gotNGN.UpdateBuyRate(decimal.NewFromFloat(10))
	gotNGN.UpdateSellRate(decimal.NewFromFloat(5))
	gotNewNGN, err := acc.UpdateCurrency(context.Background(), *gotNGN)
	assert.NoError(t, err)

	currencies, err := acc.ListCurrencies(context.Background())
	assert.NoError(t, err)
	assert.Len(t, currencies, 1)

	assert.Equal(t, gotNewNGN.Name, currencies[0].Name)
	assert.Equal(t, gotNewNGN.RateBuy.String(), currencies[0].RateBuy.String())
	assert.Equal(t, gotNewNGN.RateSell.String(), currencies[0].RateSell.String())
}
