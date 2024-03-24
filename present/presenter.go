package present

import (
	"time"

	"github.com/otyang/waas-go"
	"github.com/shopspring/decimal"
)

type (
	NewWalletResponse struct {
		ID                string            `json:"id"`
		CustomerID        string            `json:"customerId"`
		Currency          waas.Currency     `json:"currency"`
		AvailableBalance  decimal.Decimal   `json:"availableBalance"`
		LienBalance       decimal.Decimal   `json:"lienBalance"`
		TotalBalance      decimal.Decimal   `json:"totalBalance"`
		TotalBalanceInUSD decimal.Decimal   `json:"totalBalanceInUSD"`
		Status            waas.WalletStatus `json:"status"`
		CreatedAt         time.Time         `json:"createdAt"`
		UpdatedAt         time.Time         `json:"updatedAt"`
	}

	TotalBalanceResponse struct {
		CurrencyCode   string
		CurrencySymbol string
		LogoURL        string
		Total          decimal.Decimal
	}

	AllWalletsResponse struct {
		Wallets       []NewWalletResponse    `json:"wallets"`
		TotalBalances []TotalBalanceResponse `json:"totalBalances"`
	}
)

func WalletList(wallets []*waas.Wallet, currencies []waas.Currency) (*AllWalletsResponse, error) {
	var (
		usdTotalBalance decimal.Decimal
		ws              []NewWalletResponse
	)

	for _, _w := range wallets {
		_c, err := waas.FindCurrency(currencies, _w.CurrencyCode)
		if err != nil {
			return nil, err
		}

		gnwr := generateWalletResponse(_w, *_c)
		usdTotalBalance = usdTotalBalance.Add(gnwr.TotalBalanceInUSD)
		ws = append(ws, gnwr)
	}

	tb, err := calcTotalBalance(usdTotalBalance, currencies)
	if err != nil {
		return nil, err
	}

	return &AllWalletsResponse{Wallets: ws, TotalBalances: tb}, nil
}

func Wallet(wallet *waas.Wallet, currencies []waas.Currency) (*AllWalletsResponse, error) {
	return WalletList([]*waas.Wallet{wallet}, currencies)
}

func generateWalletResponse(w *waas.Wallet, c waas.Currency) NewWalletResponse {
	var usdEquivalent decimal.Decimal

	if !c.RateBuy.Equal(decimal.Zero) { // since anything divide by 0 is error/panic. let's avoid it
		usdEquivalent = w.TotalBalance().Div(c.RateBuy).RoundCeil(int32(c.Precision))
	}

	return NewWalletResponse{
		ID:                w.ID,
		CustomerID:        w.CustomerID,
		Currency:          c,
		AvailableBalance:  w.AvailableBalance.RoundCeil(int32(c.Precision)),
		LienBalance:       w.LienBalance,
		TotalBalance:      w.TotalBalance(),
		TotalBalanceInUSD: usdEquivalent,
		Status:            w.Status,
		CreatedAt:         w.CreatedAt,
		UpdatedAt:         w.UpdatedAt,
	}
}

func calcTotalBalance(totalAmountUSD decimal.Decimal, currencies []waas.Currency) ([]TotalBalanceResponse, error) {
	ngn, err := waas.FindCurrency(currencies, "NGN")
	if err != nil {
		return nil, err
	}

	usd, err := waas.FindCurrency(currencies, "USD")
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
