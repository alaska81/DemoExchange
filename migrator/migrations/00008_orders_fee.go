package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(Up00008, nil)
}

func Up00008(ctx context.Context, tx *sql.Tx) error {
	query := `
		ALTER TABLE "order" ADD fee numeric(16, 8) NULL DEFAULT 0;
		ALTER TABLE "order" ADD fee_coin varchar NULL DEFAULT ''::character varying;
	`
	_, err := tx.ExecContext(ctx, query)
	return err
}
