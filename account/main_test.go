package account

import (
	"context"
	"log"
	"os"
	"strings"
	"testing"

	dbstore "github.com/otyang/go-dbstore"
	"github.com/otyang/waas-go"
	"github.com/otyang/waas-go/currency"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/uptrace/bun"
)

var (
	TestDB      *bun.DB
	test_driver = dbstore.DriverSqlite
	test_dsn    = "file::memory:?cache=shared"
)

func setUp() *bun.DB {
	ctx := context.Background()

	db, err := dbstore.NewDBConnection(test_driver, test_dsn, 1, true)
	if err != nil {
		log.Fatal(err)
	}

	err = db.ResetModel(ctx, (*waas.Transaction)(nil), (*waas.Wallet)(nil), (*currency.Currency)(nil))
	if err != nil {
		log.Fatal(err)
	}

	return db
}

func tearDown(db *bun.DB) {
	mmodels := []any{(*waas.Transaction)(nil), (*waas.Wallet)(nil), (*currency.Currency)(nil)}
	for _, model := range mmodels {
		_, err := db.NewDropTable().Model(model).Exec(context.TODO())
		if err != nil {
			log.Fatal(err)
		}
	}
}

func createTestRandomWallet(customerID, currencyCode string) *waas.Wallet {
	return waas.NewWallet(customerID, currencyCode, false)
}

func toPointer[T any](v T) *T {
	return &v
}

func TestMain(m *testing.M) {
	TestDB = setUp()
	defer tearDown(TestDB)

	exitVal := m.Run()
	os.Exit(exitVal)
}

func TestAccount_WithTxBulkUpdateWalletAndTransaction(t *testing.T) {
	t.Parallel()

	acc := &Account{
		db: TestDB,
	}

	wallet, err := acc.CreateWallet(context.Background(), createTestRandomWallet("cust_800", "ngn"))
	assert.NoError(t, err)

	got, err := acc.Credit(context.Background(), waas.CreditWalletParams{
		WalletID:  wallet.ID,
		Amount:    decimal.NewFromFloat(50),
		Fee:       decimal.Zero,
		Type:      waas.TransactionTypeDeposit,
		Narration: "deposit of funds",
	})
	assert.NoError(t, err)

	err = acc.WithTxBulkUpdateWalletAndTransaction(
		context.Background(), []*waas.Wallet{got.Wallet}, []*waas.Transaction{got.Transaction},
	)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "constraint"))
}
