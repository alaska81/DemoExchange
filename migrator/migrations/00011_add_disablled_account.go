package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(Up00011, nil)
}

func Up00011(ctx context.Context, tx *sql.Tx) error {
	query := `
		ALTER TABLE public.account ADD disabled bool DEFAULT false NULL;
	`
	_, err := tx.ExecContext(ctx, query)
	return err
}
