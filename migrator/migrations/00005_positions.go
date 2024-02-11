package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(Up00005, nil)
}

func Up00005(ctx context.Context, tx *sql.Tx) error {
	query := `
		DROP INDEX position_status_idx;
		ALTER TABLE "position" DROP COLUMN status;

		ALTER TABLE account DROP COLUMN position_type;
	`
	_, err := tx.ExecContext(ctx, query)
	return err
}
