package present

import (
	"testing"
	"time"

	"github.com/otyang/waas-go/types"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

// Define a function to create mock transactions (customize as needed)
func createMockTransaction(isDebit bool, amount, fee, balanceAfter float64) *types.Transaction {
	return &types.Transaction{
		IsDebit:      isDebit,
		Amount:       decimal.NewFromFloat(amount),
		Fee:          decimal.NewFromFloat(fee),
		BalanceAfter: decimal.NewFromFloat(balanceAfter),
		CreatedAt:    time.Now(),
	}
}

func TestCalculateAnalyticsAndTransactions(t *testing.T) {
	t.Parallel()

	expectedAnalyticsSummary := AnalyticsSummary{
		OpeningBalance:        decimal.NewFromFloat(2.00),
		ClosingBalance:        decimal.NewFromFloat(40.00),
		TotalTransactionCount: 3,
		TotalCreditCount:      1,
		TotalDebitCount:       2,
		TotalDebit:            decimal.NewFromFloat(12),
		TotalCredit:           decimal.NewFromFloat(8.0),
		TotalFee:              decimal.NewFromFloat(0.75),
	}

	expectedTxns := []TransactionStatement{
		{
			Date:        time.Now().Format(timeLayout),
			Description: "",
			Credit:      "0.00",
			Debit:       "2.00",
			Fee:         "0.25",
			Balance:     "2.00",
		},
		{
			Date:        time.Now().Format(timeLayout),
			Description: "",
			Credit:      "8.00",
			Debit:       "0.00",
			Fee:         "0.25",
			Balance:     "0.00",
		},
		{
			Date:        time.Now().Format(timeLayout),
			Description: "",
			Credit:      "0.00",
			Debit:       "10.00",
			Fee:         "0.25",
			Balance:     "40.00",
		},
	}

	transactions := []*types.Transaction{
		createMockTransaction(true, 2, 0.25, 2.00),
		createMockTransaction(false, 8.0, 0.25, 0.00),
		createMockTransaction(true, 10.0, 0.25, 40.00),
	}

	var precision int32 = 2

	gotAnalyticSummary, gotTxns := calculateAnalyticsAndTransactions(transactions, precision)
	assert.Equal(t, expectedAnalyticsSummary.OpeningBalance, gotAnalyticSummary.OpeningBalance)
	assert.Equal(t, expectedTxns, gotTxns)
}

func Test_createTransactionStatement(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		inputTx   *types.Transaction
		precision int32
		expected  TransactionStatement
	}{
		{
			name: "Debit transaction",
			inputTx: &types.Transaction{
				Type:         "DEBIT",
				Narration:    toPointer("Test Transaction"),
				IsDebit:      true,
				Amount:       decimal.NewFromInt(10),
				Fee:          decimal.NewFromFloat(0.5),
				BalanceAfter: decimal.NewFromInt(50),
				CreatedAt:    time.Date(2024, 3, 27, 0, 0, 0, 0, time.UTC),
			},
			precision: 2,
			expected: TransactionStatement{
				Date:        "27-Mar-2024",
				Description: "DEBIT Test Transaction",
				Debit:       "10.00",
				Credit:      "0.00",
				Fee:         "0.50",
				Balance:     "50.00",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := createTransactionStatement(tc.inputTx, tc.precision)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func toPointer(s string) *string {
	return &s
}

func TestFormatAmount(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		amount    decimal.Decimal
		isDebit   bool
		precision int32
		expected  string
	}{
		{
			name:      "Debit, positive amount",
			amount:    decimal.NewFromFloat(12.345),
			isDebit:   true,
			precision: 2,
			expected:  "12.35", // Note: StringFixed rounds
		},
		{
			name:      "Debit, zero amount",
			amount:    decimal.Zero,
			isDebit:   true,
			precision: 2,
			expected:  "0.00",
		},
		{
			name:      "Credit, positive amount",
			amount:    decimal.NewFromFloat(987.65),
			isDebit:   false,
			precision: 2,
			expected:  "0.00",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := formatAmount(tc.amount, tc.isDebit, tc.precision)
			assert.Equal(t, result, tc.expected)
		})
	}
}
