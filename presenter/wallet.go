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
	TotalBalance          decimal.Decimal   `json:"totalBalance"`
	IsFrozen              bool              `json:"isFrozen"`
	CreatedAt             time.Time         `json:"createdAt"`
	UpdatedAt             time.Time         `json:"updatedAt"`
	Currency              currency.Currency `json:"currency"`
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

		ws = append(ws, NewWalletResponse{
			ID:                    _w.ID,
			CustomerID:            _w.CustomerID,
			AvailableBalance:      _w.AvailableBalance.RoundCeil(int32(_c.Precision)),
			AvailableBalanceInUSD: _w.AvailableBalance.Mul(_c.RateSell).RoundCeil(int32(2)),
			IsFrozen:              _w.IsFrozen,
			CreatedAt:             _w.CreatedAt,
			UpdatedAt:             _w.UpdatedAt,
			Currency:              *_c,
		})
	}

	return &AllWalletsResponse{
		Wallets:       ws,
		TotalBalances: calcTotalBalance(usdTotalBalance, SellRatesTotalBalance{}),
	}, nil
}
