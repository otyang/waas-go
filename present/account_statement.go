package present

import (
	"github.com/otyang/waas-go"
	"github.com/shopspring/decimal"
)

type (
	Statement struct {
		AccountName          string                 `json:"accountName"`
		AccountNumber        string                 `json:"accountNumber"`
		AccountHolderAddress string                 `json:"accountHolderAddress"`
		AccountCurrency      string                 `json:"accountCurrency"`
		DatePrinted          string                 `json:"datePrinted"`
		CurrentBalance       string                 `json:"currentBalance"`
		Summary              AnalyticsSummary       `json:"summary"`
		Transactions         []StatementTransaction `json:"transactions"`
	}

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

	StatementTransaction struct {
		Date        string `json:"date"`
		Description string `json:"description"`
		Credit      string `json:"credit"`
		Debit       string `json:"debit"`
		Fee         string `json:"totalFee"`
		Balance     string `json:"balance"`
	}
)

const (
	timeLayout = "02-Jan-2006"
)

func AccountStatement(currencies []waas.Currency, wallet *waas.Wallet, transactions []*waas.Transaction) Statement {
	as, st := generateAnalyticsAndTransactions(transactions, 2)
	return Statement{
		AccountName:          "",
		AccountNumber:        wallet.ID,
		AccountHolderAddress: "",
		AccountCurrency:      wallet.CurrencyCode,
		DatePrinted:          "",
		CurrentBalance:       wallet.TotalBalance().StringFixed(2),
		Summary:              as,
		Transactions:         st,
	}
}

func generateAnalyticsAndTransactions(transactions []*waas.Transaction, precision int32) (AnalyticsSummary, []StatementTransaction) {
	var (
		statementTransactions []StatementTransaction
		analytics             AnalyticsSummary
	)

	analytics.TotalTransactionCount = len(transactions)

	for index, transaction := range transactions {
		if index == 0 {
			analytics.OpeningBalance = transaction.Amount
		}

		if (analytics.TotalTransactionCount > 0) && (index == analytics.TotalTransactionCount-1) {
			analytics.ClosingBalance = transaction.Amount
		}

		if transaction.IsDebit {
			analytics.TotalDebitCount += 1
			analytics.TotalDebit = analytics.TotalDebit.Add(transaction.Amount)
		}

		if !transaction.IsDebit {
			analytics.TotalCreditCount += 1
			analytics.TotalCredit = analytics.TotalCredit.Add(transaction.Amount)
		}

		analytics.TotalFee = analytics.TotalFee.Add(transaction.Fee)
		statementTransactions = append(statementTransactions, newStatementTransaction(transaction, precision))
	}

	return analytics, statementTransactions
}

func newStatementTransaction(transaction *waas.Transaction, precision int32) StatementTransaction {
	narration := func() string {
		if transaction.Narration == nil {
			return string(transaction.Type)
		}
		return string(transaction.Type) + " " + *transaction.Narration
	}()

	debit, credit := func() (decimal.Decimal, decimal.Decimal) {
		if transaction.IsDebit {
			return transaction.Amount, decimal.Zero
		}
		return decimal.Zero, transaction.Amount
	}()

	return StatementTransaction{
		Date:        transaction.CreatedAt.Format(timeLayout),
		Description: narration,
		Debit:       debit.StringFixedBank(precision),
		Credit:      credit.StringFixedBank(precision),
		Fee:         transaction.Fee.StringFixedBank(precision),
		Balance:     transaction.BalanceAfter.StringFixedBank(precision),
	}
}
