package types

import (
	"fmt"
	"sort"
	"time"

	"github.com/shopspring/decimal"
)

const timeLayout = "02-Jan-2006"

// AccountStatement represents a comprehensive financial statement for a wallet account
type AccountStatement struct {
	AccountName          string                 `json:"accountName"`          // Display name of the account/wallet
	AccountNumber        string                 `json:"accountNumberOrID"`    // Unique identifier for the account
	AccountHolderAddress string                 `json:"accountHolderAddress"` // Physical/digital address of account holder
	AccountCurrency      string                 `json:"accountCurrency"`      // Currency code (e.g., "USD", "EUR")
	AccountOpened        string                 `json:"accountOpened"`        // Date when account was created (formatted)
	IsAccountFrozen      bool                   `json:"isFrozen"`             // Whether account is currently frozen
	IsAccountClosed      bool                   `json:"isClosed"`             // Whether account is closed
	DatePrinted          string                 `json:"datePrinted"`          // When this statement was generated
	CurrentBalance       string                 `json:"currentBalance"`       // Formatted current available balance
	StartDate            time.Time              `json:"createdAt"`            // Start of reporting period (raw)
	EndDate              time.Time              `json:"updatedAt"`            // End of reporting period (raw)
	Summary              AnalyticsSummary       `json:"summary"`              // Financial summary for period
	Transactions         []TransactionStatement `json:"transactions"`         // Detailed transaction records
}

// AnalyticsSummary provides aggregated financial data for the statement period
type AnalyticsSummary struct {
	OpeningBalance        decimal.Decimal `json:"openingBalance"`        // Balance at period start
	ClosingBalance        decimal.Decimal `json:"closingBalance"`        // Balance at period end
	TotalTransactionCount int             `json:"totalTransactionCount"` // Number of transactions
	TotalCreditCount      int             `json:"totalCreditCount"`      // Number of credit transactions
	TotalDebitCount       int             `json:"totalDebitCount"`       // Number of debit transactions
	TotalCreditAmount     decimal.Decimal `json:"totalCredit"`           // Sum of all credits
	TotalDebitAmount      decimal.Decimal `json:"totalDebit"`            // Sum of all debits
	TotalFee              decimal.Decimal `json:"totalFee"`              // Sum of all fees charged
}

// TransactionStatement represents a single transaction line in the account statement
type TransactionStatement struct {
	Date        string `json:"date"`        // Formatted transaction date
	Description string `json:"description"` // Transaction purpose/memo
	Credit      string `json:"credit"`      // Formatted credit amount (empty if debit)
	Debit       string `json:"debit"`       // Formatted debit amount (empty if credit)
	Fee         string `json:"fee"`         // Formatted transaction fee
	Balance     string `json:"balance"`     // Formatted balance after transaction
}

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
