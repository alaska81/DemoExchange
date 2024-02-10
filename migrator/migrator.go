package migrator

import (
	"embed"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	_ "DemoExchange/migrator/migrations"
)

const path = "migrations"

//go:embed migrations/*.go
var embedMigrations embed.FS

func Migrate(pool *pgxpool.Pool) error {
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("goose set dialect: %w", err)
	}

	db := stdlib.OpenDBFromPool(pool)

	if err := goose.Up(db, path); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}

	if err := db.Close(); err != nil {
		return fmt.Errorf("db close: %w", err)
	}

	return nil
}
