package currency

import (
	"time"

	"github.com/shopspring/decimal"
)

// Currency structure
type Currency struct {
	Code                string          `json:"code"`
	Name                string          `json:"name"`
	Symbol              string          `json:"symbol"`
	IsFiat              bool            `json:"isFiat"`
	IsStableCoin        bool            `json:"isStableCoin"`
	IconURL             string          `json:"iconUrl"`
	Precision           int             `json:"precision"`
	RateBuy             decimal.Decimal `json:"rateBuy"`
	RateSell            decimal.Decimal `json:"rateSell"`
	Disabled            bool            `json:"disabled"`
	SupportsSwap        bool            `json:"supportsSwap"`
	SupportsDeposits    bool            `json:"supportsDeposits"`
	SupportsWithdrawals bool            `json:"supportsWithdrawals"`
	DepositFee          decimal.Decimal `json:"depositFee"`
	WithdrawalFee       decimal.Decimal `json:"withdrawalFee"`
	CreatedAt           time.Time       `json:"createdAt"`
	UpdatedAt           time.Time       `json:"updatedAt"`
}
