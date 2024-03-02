package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(Up00009, nil)
}

func Up00009(ctx context.Context, tx *sql.Tx) error {
	query := `
		ALTER TABLE "order" ADD reduce_only bool NULL DEFAULT false;
	`
	_, err := tx.ExecContext(ctx, query)
	return err
}
