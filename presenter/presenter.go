package presenter

import (
	"time"

	"github.com/otyang/waas-go"
	"github.com/otyang/waas-go/currency"
	"github.com/shopspring/decimal"
)

type NewWalletResponse struct {
	ID                    string            `json:"id"`
	CustomerID            string            `json:"customerId"`
	AvailableBalance      decimal.Decimal   `json:"availableBalance"`
	AvailableBalanceInUSD decimal.Decimal   `json:"availableBalanceInUSD"`
	IsFrozen              bool              `json:"isFrozen"`
	CreatedAt             time.Time         `json:"createdAt"`
	UpdatedAt             time.Time         `json:"updatedAt"`
	Currency              currency.Currency `json:"currency"`
}

type TotalBalanceResponse struct {
	CurrencyCode   string
	CurrencySymbol string
	LogoURL        string
	Total          decimal.Decimal
}

type AllWalletsResponse struct {
	Wallets       []NewWalletResponse    `json:"wallets"`
	TotalBalances []TotalBalanceResponse `json:"totalBalances"`
}

func Wallets(wallets []*waas.Wallet, currencies []currency.Currency) (*AllWalletsResponse, error) {
	var (
		usdTotalBalance decimal.Decimal
		ws              []NewWalletResponse
	)

	for _, _w := range wallets {
		_c, err := currency.FindCurrency(currencies, _w.CurrencyCode)
		if err != nil {
			return nil, err
		}

		gnwr := generateWalletResponse(_w, *_c)
		usdTotalBalance = usdTotalBalance.Add(gnwr.AvailableBalanceInUSD)
		ws = append(ws, gnwr)
	}

	tb, err := calcTotalBalance(usdTotalBalance, currencies)
	if err != nil {
		return nil, err
	}

	return &AllWalletsResponse{Wallets: ws, TotalBalances: tb}, nil
}

func generateWalletResponse(w *waas.Wallet, c currency.Currency) NewWalletResponse {
	var usdEquivalent decimal.Decimal

	if !c.RateBuy.Equal(decimal.Zero) { // since anything divide by 0 is error/panic. let's avoid it
		usdEquivalent = w.AvailableBalance.Div(c.RateBuy).RoundCeil(int32(c.Precision))
	}

	return NewWalletResponse{
		ID:                    w.ID,
		CustomerID:            w.CustomerID,
		AvailableBalance:      w.AvailableBalance.RoundCeil(int32(c.Precision)),
		AvailableBalanceInUSD: usdEquivalent,
		IsFrozen:              w.IsFrozen,
		CreatedAt:             w.CreatedAt,
		UpdatedAt:             w.UpdatedAt,
		Currency:              c,
	}
}

func calcTotalBalance(totalAmountUSD decimal.Decimal, currencies []currency.Currency) ([]TotalBalanceResponse, error) {
	ngn, err := currency.FindCurrency(currencies, "NGN")
	if err != nil {
		return nil, err
	}

	usd, err := currency.FindCurrency(currencies, "USD")
	if err != nil {
		return nil, err
	}

	return []TotalBalanceResponse{
		{
			CurrencyCode:   ngn.Code,
			CurrencySymbol: ngn.Symbol,
			LogoURL:        ngn.IconURL,
			Total:          totalAmountUSD.Mul(ngn.RateSell).RoundCeil(int32(ngn.Precision)),
		},
		{
			CurrencyCode:   usd.Code,
			CurrencySymbol: usd.Symbol,
			LogoURL:        usd.IconURL,
			Total:          totalAmountUSD.Mul(usd.RateSell).RoundCeil(int32(usd.Precision)),
		},
	}, nil
}
