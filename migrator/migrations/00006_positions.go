package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(Up00006, nil)
}

func Up00006(ctx context.Context, tx *sql.Tx) error {
	query := `
		ALTER TABLE "position" ADD margin numeric(16, 8) NULL DEFAULT 0;
	`
	_, err := tx.ExecContext(ctx, query)
	return err
}
