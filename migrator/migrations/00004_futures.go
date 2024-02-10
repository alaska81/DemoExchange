package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(Up00004, nil)
}

func Up00004(ctx context.Context, tx *sql.Tx) error {
	query := `
		ALTER TABLE wallet ADD update_ts int8 NULL;
		ALTER TABLE "position" DROP COLUMN hold_coin;
		ALTER TABLE "order" ADD position_side varchar NULL DEFAULT 'both';
		ALTER TABLE "order" DROP COLUMN hold_amount;
		ALTER TABLE "order" DROP COLUMN hold_coin;
	`
	_, err := tx.ExecContext(ctx, query)
	return err
}
