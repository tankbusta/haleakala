package database

import (
	"context"
	"database/sql"

	"github.com/rs/zerolog/log"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
	"github.com/uptrace/bun/migrate"
)

var Migrations = migrate.NewMigrations()

func GetDatabase(connection string) (*bun.DB, error) {
	sqldb, err := sql.Open(sqliteshim.ShimName, connection)
	if err != nil {
		return nil, err
	}

	db := bun.NewDB(sqldb, sqlitedialect.New())
	if err := db.Ping(); err != nil {
		return nil, err
	}

	mr := migrate.NewMigrator(db, Migrations)
	ctx := context.Background()

	if err := mr.Init(ctx); err != nil {
		return nil, err
	}

	if _, err := mr.Migrate(ctx); err != nil {
		return nil, err
	}

	if _, err := db.Exec("PRAGMA journal_mode=WAL;"); err != nil {
		log.Error().Err(err).Msg("failed to set journal mode")
	}

	return db, nil
}
