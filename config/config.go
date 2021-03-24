// Config is put into a different package to prevent cyclic imports in case
// it is needed in several locations

package config

import (
	"time"
)

type Config struct {
	// db flavor. defaults to postgresql
	DBConfig DB `config:"db_config"`
}

// Config holds configuration for database to watch.
type DB struct {
	// URI holds URI information for database
	URI       string `config:"uri"`
	minReconn int    `config:"min_reconn"`
	maxReconn int    `config:"max_reconn"`
	MinReconn time.Duration
	MaxReconn time.Duration
	// WatchAll defines whether all CRUD from all tables are watched.
	// if true, TableOperations is ignored
	// if false, TableOperations must not be nil.
	WatchAll bool `config:"watch_all"`
	// TableOps define which table and CRUD operations to watch.
	TableOps map[string][]string `config:"table_ops"`
}

var DefaultConfig = Config{
	DBConfig: DB{
		URI:       "postgres://postgres:pwd@localhost:5432?sslmode=disable",
		MinReconn: 10 * time.Second,
		MaxReconn: 1 * time.Minute,
		WatchAll:  true,
	},
}
