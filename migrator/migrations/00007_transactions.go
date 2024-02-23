package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(Up00007, nil)
}

func Up00007(ctx context.Context, tx *sql.Tx) error {
	query := `
		CREATE TABLE "transaction" (
			transaction_uid uuid NOT NULL,
			exchange varchar NOT NULL,
			account_uid uuid NOT NULL,
			symbol varchar NOT NULL,
			transaction_type varchar NOT NULL,
			amount numeric(16, 8) NOT NULL,
			create_ts int8 NOT NULL,
			CONSTRAINT transaction_pk PRIMARY KEY (transaction_uid)
		);

		ALTER TABLE "transaction" ADD CONSTRAINT transaction_account_fk FOREIGN KEY (account_uid) REFERENCES account(account_uid) ON DELETE CASCADE;
		CREATE INDEX transaction_account_uid_idx ON "transaction" (account_uid);
		CREATE INDEX transaction_exchange_idx ON "transaction" (exchange);

		CREATE INDEX position_exchange_idx ON "position" (exchange);
	`
	_, err := tx.ExecContext(ctx, query)
	return err
}
