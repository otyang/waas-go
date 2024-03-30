package present

import (
	"time"

	"github.com/otyang/waas-go"
	"github.com/shopspring/decimal"
)

type NewWalletResponse struct {
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

type TotalBalanceResponse struct {
	CurrencyCode   string
	CurrencySymbol string
	LogoURL        string
	Total          string
}

func WalletList(wallets []*waas.Wallet) ([]NewWalletResponse, decimal.Decimal, error) {
	var (
		allWalletsBalanceInUSD decimal.Decimal
		walletResponses        []NewWalletResponse
	)

	for _, w := range wallets {
		var usdEquivalent decimal.Decimal

		if !w.Currency.RateBuy.Equal(decimal.Zero) {
			// anything divide by 0 is error/panic. let's avoid it
			usdEquivalent = w.TotalBalance().Div(w.Currency.RateBuy).RoundBank(int32(w.Currency.Precision))
			allWalletsBalanceInUSD = allWalletsBalanceInUSD.Add(usdEquivalent)
		}

		walletResponses = append(walletResponses, NewWalletResponse{
			ID:                w.ID,
			CustomerID:        w.CustomerID,
			Currency:          w.Currency,
			AvailableBalance:  w.AvailableBalance.RoundBank(int32(w.Currency.Precision)),
			LienBalance:       w.LienBalance.RoundBank(int32(w.Currency.Precision)),
			TotalBalance:      w.TotalBalance().RoundBank(int32(w.Currency.Precision)),
			TotalBalanceInUSD: usdEquivalent,
			Status:            w.Status,
			CreatedAt:         w.CreatedAt,
			UpdatedAt:         w.UpdatedAt,
		})
	}

	return walletResponses, allWalletsBalanceInUSD, nil
}

func Wallet(wallet *waas.Wallet, currencies []waas.Currency) (*NewWalletResponse, error) {
	response, _, err := WalletList([]*waas.Wallet{wallet})
	if err != nil {
		return nil, err
	}
	return &response[0], nil
}

func TotalBalances(totalAmountUSD decimal.Decimal, currencies []waas.Currency) ([]TotalBalanceResponse, error) {
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
			Total:          totalAmountUSD.Mul(ngn.RateSell).StringFixedBank(int32(ngn.Precision)),
		},
		{
			CurrencyCode:   usd.Code,
			CurrencySymbol: usd.Symbol,
			LogoURL:        usd.IconURL,
			Total:          totalAmountUSD.Mul(usd.RateSell).StringFixedBank(int32(usd.Precision)),
		},
	}, nil
}
