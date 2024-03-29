package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(Up00010, nil)
}

func Up00010(ctx context.Context, tx *sql.Tx) error {
	query := `
		ALTER TABLE "order" ADD leverage int2 NULL DEFAULT 1;
	`
	_, err := tx.ExecContext(ctx, query)
	return err
}
