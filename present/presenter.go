package present

import (
	"time"

	"github.com/otyang/waas-go/types"
	"github.com/shopspring/decimal"
)

type AllWalletsAndTotal struct {
	OverallTotalUSDBalance decimal.Decimal
	NewWalletsResponses    []NewWalletResponse
}

type NewWalletResponse struct {
	ID                string          `json:"id"`
	CustomerID        string          `json:"customerId"`
	Currency          types.Currency  `json:"currency"`
	AvailableBalance  decimal.Decimal `json:"availableBalance"`
	LienBalance       decimal.Decimal `json:"lienBalance"`
	TotalBalance      decimal.Decimal `json:"totalBalance"`
	TotalBalanceInUSD decimal.Decimal `json:"totalBalanceInUSD"`
	IsFrozen          bool            `json:"isFrozen"`
	IsClosed          bool            `json:"isClosed"`
	CreatedAt         time.Time       `json:"createdAt"`
	UpdatedAt         time.Time       `json:"updatedAt"`
}

type TotalBalanceResponse struct {
	CurrencyCode   string
	CurrencySymbol string
	LogoURL        string
	Total          string
}

func WalletList(wallets []*types.Wallet, currencies []types.Currency) (*AllWalletsAndTotal, error) {
	var (
		allWalletsBalanceInUSD decimal.Decimal
		walletResponses        []NewWalletResponse
	)

	for _, w := range wallets {
		var usdEquivalent decimal.Decimal

		wCurrency, err := types.FindCurrency(currencies, w.CurrencyCode)
		if err != nil {
			return nil, err
		}

		if !wCurrency.RateBuy.Equal(decimal.Zero) {
			// anything divide by 0 is error/panic. let's avoid it
			usdEquivalent = w.TotalBalance().Div(wCurrency.RateBuy).RoundBank(int32(wCurrency.Precision))
		}

		allWalletsBalanceInUSD = allWalletsBalanceInUSD.Add(usdEquivalent)
		walletResponses = append(walletResponses, NewWalletResponse{
			ID:                w.ID,
			CustomerID:        w.CustomerID,
			Currency:          *wCurrency,
			AvailableBalance:  w.AvailableBalance.RoundBank(int32(wCurrency.Precision)),
			LienBalance:       w.LienBalance.RoundBank(int32(wCurrency.Precision)),
			TotalBalance:      w.TotalBalance().RoundBank(int32(wCurrency.Precision)),
			TotalBalanceInUSD: usdEquivalent,
			IsFrozen:          w.IsFrozen,
			IsClosed:          w.IsClosed,
			CreatedAt:         w.CreatedAt,
			UpdatedAt:         w.UpdatedAt,
		})
	}

	return &AllWalletsAndTotal{
		OverallTotalUSDBalance: allWalletsBalanceInUSD,
		NewWalletsResponses:    walletResponses,
	}, nil
}

func Wallet(wallet *types.Wallet, currencies []types.Currency) (*NewWalletResponse, error) {
	response, err := WalletList([]*types.Wallet{wallet}, currencies)
	if err != nil {
		return nil, err
	}
	return &response.NewWalletsResponses[0], nil
}

func TotalBalances(totalAmountUSD decimal.Decimal, currencies []types.Currency) ([]TotalBalanceResponse, error) {
	ngn, err := types.FindCurrency(currencies, "NGN")
	if err != nil {
		return nil, err
	}

	usd, err := types.FindCurrency(currencies, "USD")
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
