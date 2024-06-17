package presenter

import (
	"strings"
	"testing"
	"time"

	"github.com/otyang/waas-go/types"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func toPointer(s string) *string {
	return &s
}

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
			Description: " ",
			Credit:      "",
			Debit:       "2.00",
			Fee:         "0.25",
			Balance:     "2.00",
		},
		{
			Date:        time.Now().Format(timeLayout),
			Description: " ",
			Credit:      "8.00",
			Debit:       "",
			Fee:         "0.45",
			Balance:     "0.00",
		},
		{
			Date:        time.Now().Format(timeLayout),
			Description: " ",
			Credit:      "",
			Debit:       "10.00",
			Fee:         "0.65",
			Balance:     "40.00",
		},
	}

	transactions := []*types.Transaction{
		createMockTransaction(true, 2, 0.25, 2.00),
		createMockTransaction(false, 8.0, 0.45, 0.00),
		createMockTransaction(true, 10.0, 0.65, 40.00),
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
				Credit:      "",
				Fee:         "0.50",
				Balance:     "50.00",
			},
		},
		{
			name: "Credit transaction",
			inputTx: &types.Transaction{
				Type:         "CREDIT",
				Narration:    toPointer("Test Transaction"),
				IsDebit:      false,
				Amount:       decimal.NewFromInt(20),
				Fee:          decimal.NewFromFloat(0.5),
				BalanceAfter: decimal.NewFromInt(100),
				CreatedAt:    time.Date(2024, 3, 27, 0, 0, 0, 0, time.UTC),
			},
			precision: 2,
			expected: TransactionStatement{
				Date:        "27-Mar-2024",
				Description: "CREDIT Test Transaction",
				Debit:       "",
				Credit:      "20.00",
				Fee:         "0.50",
				Balance:     "100.00",
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

func TestFormatAmountDebit(t *testing.T) {
	amount := decimal.NewFromFloat(123.45)
	precision := int32(2)
	creditAmount, debitAmount := formatAmount(amount, true, precision)

	assert.Equal(t, "", creditAmount)
	assert.Equal(t, amount.String(), debitAmount)
}

func TestFormatAmountCredit(t *testing.T) {
	amount := decimal.NewFromFloat(123.45)
	precision := int32(2)
	creditAmount, debitAmount := formatAmount(amount, false, precision)

	assert.Equal(t, amount.String(), creditAmount)
	assert.Equal(t, "", debitAmount)
}

func TestGetNarrationWithNarration(t *testing.T) {
	tx := &types.Transaction{
		Type:      types.TransactionType("swap"),
		Narration: toPointer("Some narration"),
	}

	expectedNarration := string(tx.Type) + " " + "Some narration"
	actualNarration := getNarration(tx)

	assert.True(t, strings.EqualFold(expectedNarration, actualNarration))
}

func TestGetNarrationWithoutNarration(t *testing.T) {
	tx := &types.Transaction{
		Type: types.TransactionType("withdrawal"),
	}

	expectedNarration := string(tx.Type) + " "
	actualNarration := getNarration(tx)

	assert.True(t, strings.EqualFold(expectedNarration, actualNarration))
}
