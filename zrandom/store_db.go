package zrandom

// store

import (
	"database/sql"
	"fmt"
	"runtime"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
	"github.com/uptrace/bun/extra/bundebug"
)

// DBConfig holds configuration for database connection
type DBConfig struct {
	Driver         string // Database driver ("postgres" or "sqlite3")
	DataSourceName string // Connection string
	PoolMax        int    // Maximum number of connections in the pool (per CPU core)
	PrintQueries   bool   // Whether to enable query logging
}

// NewDBConnection establishes a connection to the database and returns both the raw sql.DB and bun.DB connections.
func NewDBConnection(cfg DBConfig) (*sql.DB, *bun.DB, error) {
	// Validate driver type
	var sqlDriver string
	switch cfg.Driver {
	case "postgres":
		sqlDriver = "postgres"
	case "sqlite", "sqlite3":
		sqlDriver = sqliteshim.ShimName
	default:
		return nil, nil, fmt.Errorf("unsupported database driver: %q (must be either 'postgres' or 'sqlite3')", cfg.Driver)
	}

	// Open raw database connection
	sqlDB, err := sql.Open(sqlDriver, cfg.DataSourceName)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Verify the connection is alive
	if err := sqlDB.Ping(); err != nil {
		sqlDB.Close()
		return nil, nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Configure connection pool
	maxProcs := runtime.GOMAXPROCS(0)
	sqlDB.SetMaxOpenConns(cfg.PoolMax * maxProcs)
	sqlDB.SetMaxIdleConns(cfg.PoolMax * maxProcs)

	// Create bun.DB wrapper with the appropriate dialect
	var bunDB *bun.DB
	switch cfg.Driver {
	case "postgres":
		bunDB = bun.NewDB(sqlDB, pgdialect.New())
	case "sqlite", "sqlite3":
		bunDB = bun.NewDB(sqlDB, sqlitedialect.New())
	}

	// Configure query logging if enabled
	if cfg.PrintQueries {
		queryHook := bundebug.NewQueryHook(
			bundebug.WithVerbose(true),
			bundebug.WithEnabled(true),
		)
		bunDB.AddQueryHook(queryHook)
	}

	return sqlDB, bunDB, nil
}
