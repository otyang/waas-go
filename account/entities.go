package account

type (
	ListWalletsFilterOpts struct {
		CustomerID    string
		CurrencyCodes []string
		IsFrozen      *bool
		IsClosed      *bool
	}
)
