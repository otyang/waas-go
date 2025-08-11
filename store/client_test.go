package store

import (
	"context"
	"log"

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

	if _, err = dbstore.NewWithMigration(db); err != nil {
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
