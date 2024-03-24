package waas

import (
	"time"

	"github.com/shopspring/decimal"
)

type (
	// CreditWalletParams defines parameters for crediting a wallet.
	CreditWalletParams struct {
		WalletID  string
		Amount    decimal.Decimal
		Fee       decimal.Decimal
		Type      TransactionType
		Narration string
	}

	CreditWalletResponse struct {
		Wallet      *Wallet
		Transaction *Transaction
	}

	// DebitWalletParams defines parameters for debiting a wallet.
	DebitWalletParams struct {
		WalletID  string
		Amount    decimal.Decimal
		Fee       decimal.Decimal
		Type      TransactionType
		Status    TransactionStatus
		Narration string
	}

	DebitWalletResponse struct {
		Wallet      *Wallet
		Transaction *Transaction
	}

	// TransferParams defines parameters for transferring funds between wallets.
	TransferRequestParams struct {
		FromWalletID string          `json:"fromWid"`
		ToWalletID   string          `json:"toWid"`
		Amount       decimal.Decimal `json:"amount"`
		Fee          decimal.Decimal `json:"fee"`
		Narration    string          `json:"narration"`
	}

	TransferResponse struct {
		FromWallet      *Wallet
		ToWallet        *Wallet
		FromTransaction *Transaction
		ToTransaction   *Transaction
	}

	// ReverseParams defines parameters for reversing a transaction.
	ReverseRequestParams struct {
		TransactionID string `json:"transactionId"`
	}

	ReverseResponse struct {
		OldTx  *Transaction
		NewTx  *Transaction
		Wallet *Wallet
	}

	// SwapParams defines parameters for swapping currencies between wallets.
	SwapRequestParams struct {
		UserID           string
		FromCurrencyCode string
		ToCurrencyCode   string
		FromAmount       decimal.Decimal
		FromFee          decimal.Decimal
		ToAmount         decimal.Decimal
	}

	SwapWalletResponse struct {
		FromWallet      *Wallet
		ToWallet        *Wallet
		FromTransaction *Transaction
		ToTransaction   *Transaction
	}

	ListWalletsFilterParams struct {
		CustomerID    *string
		CurrencyCodes []string
		Status        *WalletStatus
	}

	ListTransactionsFilterParams struct {
		Limit      int
		Before     time.Time
		After      time.Time
		CustomerID *string
		WalletID   *string
		Currency   []string
		IsDebit    *bool
		Type       *TransactionType
		Status     *TransactionStatus
		Reversed   *bool
	}

	ListOrder string
)

const (
	Asc  ListOrder = "asc"
	Desc ListOrder = "desc"
)
