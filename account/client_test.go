package account

import (
	"context"
	"log"
	"os"
	"testing"

	dbstore "github.com/otyang/go-dbstore"
	"github.com/otyang/waas-go/types"
	"github.com/uptrace/bun"
)

var (
	TestDB      *bun.DB
	test_driver = dbstore.DriverSqlite
	test_dsn    = "file::memory:?cache=shared"
)

func setUp() *bun.DB {
	db, err := dbstore.NewDBConnection(test_driver, test_dsn, 1, true)
	if err != nil {
		log.Fatal(err)
	}

	if _, err = NewWithMigration(db); err != nil {
		log.Fatal(err)
	}

	return db
}

func tearDown(db *bun.DB) {
	mmodels := []any{(*types.Transaction)(nil), (*types.Wallet)(nil), (*types.Currency)(nil)}
	for _, model := range mmodels {
		_, err := db.NewDropTable().Model(model).Exec(context.TODO())
		if err != nil {
			log.Fatal(err)
		}
	}
}

func createTestRandomWallet(customerID, currencyCode string) *types.Wallet {
	return types.NewWallet(customerID, currencyCode)
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
