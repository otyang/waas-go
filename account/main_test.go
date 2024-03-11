package account

import (
	"context"
	"log"
	"os"
	"testing"

	dbstore "github.com/otyang/go-dbstore"
	"github.com/otyang/waas-go"
	"github.com/otyang/waas-go/currency"
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

func TestMain(m *testing.M) {
	TestDB = setUp()
	defer tearDown(TestDB)

	exitVal := m.Run()
	os.Exit(exitVal)
}
