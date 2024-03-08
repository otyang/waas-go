package currency

import (
	"time"

	"github.com/shopspring/decimal"
)

// Currency structure
type Currency struct {
	Code          string          `json:"code"`
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
