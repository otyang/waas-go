package waas

import "github.com/shopspring/decimal"

type CreditWalletResponse struct {
	AmountTransfered decimal.Decimal
	Fee              decimal.Decimal
	Wallet           Wallet
	Transaction      Transaction
}

type DebitWalletResponse struct {
	AmountTransfered decimal.Decimal
	Fee              decimal.Decimal
	Wallet           Wallet
	Transaction      Transaction
}

type SwapWalletResponse struct {
	FromWallet      Wallet
	ToWallet        Wallet
	FromTransaction Transaction
	ToTransaction   Transaction
}

type TransferResponse struct {
	AmountTransfered decimal.Decimal
	Fee              decimal.Decimal
	FromWallet       Wallet
	ToWallet         Wallet
	FromTransaction  Transaction
	ToTransaction    Transaction
}

type ReverseResponse struct {
	OldUpdatedTx  *Transaction
	NewTx         *Transaction
	UpdatedWallet *Wallet
}

// CreditWalletParams defines parameters for crediting a wallet.
type CreditWalletParams struct {
	WalletID        string
	Amount          decimal.Decimal
	Fee             decimal.Decimal
	Type            TransactionType
	SourceNarration string
}

// DebitWalletParams defines parameters for debiting a wallet.
type DebitWalletParams struct {
	WalletID        string
	Amount          decimal.Decimal
	Fee             decimal.Decimal
	Type            TransactionType
	SourceNarration string
}

// TransferParams defines parameters for transferring funds between wallets.
type TransferRequestParams struct {
	FromWalletID    string          `json:"fromWid"`
	ToWalletID      string          `json:"toWid"`
	Amount          decimal.Decimal `json:"amount"`
	Fee             decimal.Decimal `json:"fee"`
	SourceNarration string
}

// ReverseParams defines parameters for reversing a transaction.
type ReverseRequestParams struct {
	TransactionID string `json:"transactionId"`
}

// SwapParams defines parameters for swapping currencies between wallets.
type SwapRequestParams struct {
	UserID           string
	FromCurrencyCode string
	ToCurrencyCode   string
	Amount           decimal.Decimal
	Fee              decimal.Decimal
}

type ListWalletsFilterParams struct {
	CustomerID   *string
	CurrencyCode *string
	IsFiat       *bool
	IsFrozen     bool
}

type ListTransactionsFilterParams struct {
	CustomerID string
	Currency   string
	IsDebit    *bool
	Type       *TransactionType
	Status     *TransactionStatus
	Narration  string
	Reversed   *bool
}
