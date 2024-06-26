package presenter

import (
	"fmt"
	"time"

	"github.com/otyang/waas-go/types"
	"github.com/shopspring/decimal"
)

const timeLayout = "02-Jan-2006"

type AccountStatement struct {
	AccountName          string                 `json:"accountName"`
	AccountNumberOrID    string                 `json:"accountNumberOrID"`
	AccountHolderAddress string                 `json:"accountHolderAddress"`
	AccountCurrency      string                 `json:"accountCurrency"`
	AccountOpened        string                 `json:"accountOpened"`
	IsAccountFrozen      bool                   `json:"isFrozen"`
	IsAccountClosed      bool                   `json:"isClosed"`
	DatePrinted          string                 `json:"datePrinted"`
	CurrentBalance       string                 `json:"currentBalance"`
	Summary              AnalyticsSummary       `json:"summary"`
	Transactions         []TransactionStatement `json:"transactions"`
}

type (
	AnalyticsSummary struct {
		OpeningBalance        decimal.Decimal `json:"openingBalance"`
		ClosingBalance        decimal.Decimal `json:"closingBalance"`
		TotalTransactionCount int             `json:"totalTransactionCount"`
		TotalCreditCount      int             `json:"totalCreditCount"`
		TotalDebitCount       int             `json:"totalDebitCount"`
		TotalCredit           decimal.Decimal `json:"totalCredit"`
		TotalDebit            decimal.Decimal `json:"totalDebit"`
		TotalFee              decimal.Decimal `json:"totalFee"`
	}

	TransactionStatement struct {
		Date        string `json:"date"`
		Description string `json:"description"`
		Credit      string `json:"credit"`
		Debit       string `json:"debit"`
		Fee         string `json:"fee"`
		Balance     string `json:"balance"`
	}
)

func (p *Client) AccountStatement(
	wallet *types.Wallet,
	transactions []*types.Transaction,
	currencyName string,
	currencyPrecision int,
) AccountStatement {
	accMetrics, listOfTxns := calculateAnalyticsAndTransactions(transactions, int32(currencyPrecision))
	return AccountStatement{
		AccountName:          "{account name}",
		AccountNumberOrID:    wallet.ID,
		AccountHolderAddress: "{account holder}",
		AccountCurrency:      currencyName,
		AccountOpened:        time.Now().Format(timeLayout),
		IsAccountFrozen:      wallet.IsFrozen,
		IsAccountClosed:      wallet.IsClosed,
		DatePrinted:          time.Now().Format(timeLayout),
		CurrentBalance:       wallet.TotalBalance().StringFixed(2),
		Summary:              accMetrics,
		Transactions:         listOfTxns,
	}
}

func calculateAnalyticsAndTransactions(transactions []*types.Transaction, precision int32) (AnalyticsSummary, []TransactionStatement) {
	var (
		analytics             AnalyticsSummary
		statementTransactions []TransactionStatement
	)

	analytics.TotalTransactionCount = len(transactions)

	for index, tx := range transactions {
		if index == 0 {
			fmt.Println(tx.BalanceAfter)
			analytics.OpeningBalance = tx.BalanceAfter
		}

		if (analytics.TotalTransactionCount > 0) && (index == analytics.TotalTransactionCount-1) {
			analytics.ClosingBalance = tx.BalanceAfter
		}

		if tx.IsDebit {
			analytics.TotalDebitCount++
			analytics.TotalDebit = analytics.TotalDebit.Add(tx.Amount)
		} else {
			analytics.TotalCreditCount++
			analytics.TotalCredit = analytics.TotalCredit.Add(tx.Amount)
		}

		analytics.TotalFee = analytics.TotalFee.Add(tx.Fee)
		statementTransactions = append(statementTransactions, createTransactionStatement(tx, precision))
	}

	return analytics, statementTransactions
}

// createTransactionStatement formats a transaction for the statement.
func createTransactionStatement(tx *types.Transaction, precision int32) TransactionStatement {
	creditAmount, debitAmount := formatAmount(tx.Amount, tx.IsDebit, precision)
	return TransactionStatement{
		Date:        tx.CreatedAt.Format(timeLayout),
		Description: getNarration(tx),
		Debit:       debitAmount,
		Credit:      creditAmount,
		Fee:         tx.Fee.StringFixedBank(precision),
		Balance:     tx.BalanceAfter.StringFixedBank(precision),
	}
}

// formatAmount helps apply debit/credit formatting
func formatAmount(amount decimal.Decimal, isDebit bool, precision int32) (creditAmount string, debitAmount string) {
	if isDebit {
		return "", amount.StringFixed(precision)
	}
	return amount.StringFixed(precision), ""
}

func getNarration(tx *types.Transaction) string {
	if tx.Narration != nil {
		return fmt.Sprintf("%s %s", string(tx.Type), *tx.Narration)
	}
	return fmt.Sprintf("%s ", string(tx.Type))
}
