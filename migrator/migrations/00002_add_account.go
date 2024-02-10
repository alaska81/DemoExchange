package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(Up00002, nil)
}

func Up00002(ctx context.Context, tx *sql.Tx) error {
	query := `
		ALTER TABLE "position" ADD position_type varchar NOT NULL;

		CREATE TABLE account (
			account_uid uuid NOT NULL,
			service varchar NOT NULL,
			user_id varchar NOT NULL,
			position_mode varchar NOT NULL DEFAULT 'oneway'::character varying,
			position_type varchar NOT NULL DEFAULT 'isolated'::character varying,
			create_ts int8 NOT NULL,
			update_ts int8 NOT NULL,
			CONSTRAINT account_pk PRIMARY KEY (account_uid)
		);

		ALTER TABLE apikey DROP COLUMN service;
		ALTER TABLE apikey DROP COLUMN account_id;
		ALTER TABLE apikey ADD disabled bool NULL DEFAULT false;
		ALTER TABLE apikey ADD update_ts int8 NULL;
	`
	_, err := tx.ExecContext(ctx, query)
	return err
}
