package types

import (
	"fmt"
	"sort"
	"time"

	"github.com/shopspring/decimal"
)

const timeLayout = "02-Jan-2006"

type (
	AccountStatement struct {
		AccountName          string `json:"accountName"`
		AccountNumber        string `json:"accountNumberOrID"`
		AccountHolderAddress string `json:"accountHolderAddress"`
		AccountCurrency      string `json:"accountCurrency"`
		AccountOpened        string `json:"accountOpened"`
		IsAccountFrozen      bool   `json:"isFrozen"`
		IsAccountClosed      bool   `json:"isClosed"`
		DatePrinted          string `json:"datePrinted"`
		CurrentBalance       string `json:"currentBalance"`
		StartDate            time.Time
		EndDate              time.Time
		Summary              AnalyticsSummary       `json:"summary"`
		Transactions         []TransactionStatement `json:"transactions"`
	}

	AnalyticsSummary struct {
		OpeningBalance        decimal.Decimal `json:"openingBalance"`
		ClosingBalance        decimal.Decimal `json:"closingBalance"`
		TotalTransactionCount int             `json:"totalTransactionCount"`
		TotalCreditCount      int             `json:"totalCreditCount"`
		TotalDebitCount       int             `json:"totalDebitCount"`
		TotalCreditAmount     decimal.Decimal `json:"totalCredit"`
		TotalDebitAmount      decimal.Decimal `json:"totalDebit"`
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

// GenerateAccountStatement creates a comprehensive account statement for a given wallet
// including transaction history and analytics summary within a specified date range.
//
// Parameters:
//   - wallet: Pointer to the Wallet struct containing account information
//   - transactions: Slice of TransactionHistory records for the wallet
//   - startDate: Beginning of the reporting period (inclusive)
//   - endDate: End of the reporting period (inclusive)
//   - precision: Number of decimal places to round monetary values to
//
// Returns:
//   - AccountStatement containing formatted statement data
func GenerateAccountStatement(wallet *Wallet, transactions []TransactionHistory,
	startDate, endDate time.Time, precision int32,
) AccountStatement {
	// Helper function to format decimal values with consistent precision
	formatDecimal := func(d decimal.Decimal) string {
		return d.Round(precision).StringFixed(precision)
	}

	// Initialize summary with zero values properly formatted
	summary := AnalyticsSummary{
		OpeningBalance:    decimal.Zero,
		ClosingBalance:    decimal.Zero,
		TotalCreditAmount: decimal.Zero,
		TotalDebitAmount:  decimal.Zero,
		TotalFee:          decimal.Zero,
	}

	// Filter and sort transactions chronologically
	var filteredTx []TransactionHistory
	for _, tx := range transactions {
		// Include transactions within date range (inclusive)
		if (tx.CreatedAt.Equal(startDate) || tx.CreatedAt.After(startDate)) &&
			(tx.CreatedAt.Equal(endDate) || tx.CreatedAt.Before(endDate)) {
			filteredTx = append(filteredTx, tx)
		}
	}

	// Sort by initiation time (oldest first)
	sort.Slice(filteredTx, func(i, j int) bool {
		return filteredTx[i].CreatedAt.Before(filteredTx[j].CreatedAt)
	})

	// Process transactions to build statement lines and summary
	var transactionStatements []TransactionStatement
	for i, tx := range filteredTx {
		// Set opening balance from first transaction's balance before
		if i == 0 {
			summary.OpeningBalance = tx.BalanceBefore
		}

		// Update transaction counters
		summary.TotalTransactionCount++

		// Update credit/debit totals
		switch tx.Type {
		case TypeCredit:
			summary.TotalCreditCount++
			summary.TotalCreditAmount = summary.TotalCreditAmount.Add(tx.Amount)
		case TypeDebit:
			summary.TotalDebitCount++
			summary.TotalDebitAmount = summary.TotalDebitAmount.Add(tx.Amount)
		}

		// Accumulate fees
		summary.TotalFee = summary.TotalFee.Add(tx.Fee)

		// Always update closing balance to current transaction's balance after
		summary.ClosingBalance = tx.BalanceAfter

		// Prepare statement line
		stmt := TransactionStatement{
			Date:        tx.CreatedAt.Format(timeLayout), // Use consistent date format
			Description: tx.Description,
			Balance:     formatDecimal(tx.BalanceAfter),
			Fee:         formatDecimal(tx.Fee),
		}

		// Set credit/debit amounts (show zero for opposite type)
		if tx.Type == TypeCredit {
			stmt.Credit = formatDecimal(tx.Amount)
			stmt.Debit = formatDecimal(decimal.Zero)
		} else {
			stmt.Credit = formatDecimal(decimal.Zero)
			stmt.Debit = formatDecimal(tx.Amount)
		}

		transactionStatements = append(transactionStatements, stmt)
	}

	// Handle case with no transactions in period
	if len(filteredTx) == 0 {
		summary.OpeningBalance = wallet.AvailableBalance
		summary.ClosingBalance = wallet.AvailableBalance
	}

	// Format dates for display
	datePrinted := time.Now().Format(timeLayout)
	accountOpened := wallet.CreatedAt.Format(timeLayout)

	return AccountStatement{
		AccountName:          fmt.Sprintf("%s Wallet", wallet.CurrencyCode), // More descriptive name
		AccountNumber:        wallet.ID,
		AccountHolderAddress: "Not specified", // Default value
		AccountCurrency:      wallet.CurrencyCode,
		AccountOpened:        accountOpened,
		IsAccountFrozen:      wallet.Frozen,
		IsAccountClosed:      wallet.IsClosed,
		DatePrinted:          datePrinted,
		CurrentBalance:       formatDecimal(wallet.AvailableBalance),
		StartDate:            startDate,
		EndDate:              endDate,
		Summary:              summary,
		Transactions:         transactionStatements,
	}
}
