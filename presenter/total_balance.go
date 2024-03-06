package presenter

import (
	"github.com/shopspring/decimal"
)

type TotalBalanceResponse struct {
	CurrencyCode   string
	CurrencySymbol string
	LogoURL        string
	Total          decimal.Decimal
}

type SellRatesTotalBalance struct {
	USD decimal.Decimal
	NGN decimal.Decimal
	BTC decimal.Decimal
}

func calcTotalBalance(totalAmountUSD decimal.Decimal, rates SellRatesTotalBalance) []TotalBalanceResponse {
	return []TotalBalanceResponse{
		{
			CurrencyCode:   "NGN",
			CurrencySymbol: "₦",
			LogoURL:        "https://loalhost/ngn.jpg",
			Total:          rates.NGN.Mul(totalAmountUSD).RoundCeil(int32(2)),
		},
		{
			CurrencyCode:   "USD",
			CurrencySymbol: "$",
			LogoURL:        "https://loalhost/usd.jpg",
			Total:          rates.USD.Mul(totalAmountUSD).RoundCeil(int32(2)),
		},
	}
}
