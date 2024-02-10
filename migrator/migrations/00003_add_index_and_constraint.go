package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(Up00003, nil)
}

func Up00003(ctx context.Context, tx *sql.Tx) error {
	query := `
		ALTER TABLE apikey ADD CONSTRAINT apikey_account_fk FOREIGN KEY (account_uid) REFERENCES account(account_uid) ON DELETE CASCADE;
		ALTER TABLE wallet ADD CONSTRAINT wallet_account_fk FOREIGN KEY (account_uid) REFERENCES account(account_uid) ON DELETE CASCADE;
		ALTER TABLE "position" ADD CONSTRAINT position_account_fk FOREIGN KEY (account_uid) REFERENCES account(account_uid) ON DELETE CASCADE;
		ALTER TABLE "order" ADD CONSTRAINT order_account_fk FOREIGN KEY (account_uid) REFERENCES account(account_uid) ON DELETE CASCADE;

		CREATE INDEX account_service_user_id_idx ON account (service,user_id);
		CREATE INDEX apikey_account_uid_idx ON apikey (account_uid);
		CREATE INDEX apikey_disabled_idx ON apikey (disabled);
		CREATE INDEX position_account_uid_idx ON "position" (account_uid);
		CREATE INDEX position_symbol_idx ON "position" (symbol);
		CREATE INDEX position_side_idx ON "position" (side);
		CREATE INDEX position_status_idx ON "position" (status);
		CREATE INDEX order_account_uid_idx ON "order" (account_uid);
		CREATE INDEX order_exchange_idx ON "order" (exchange);
		CREATE INDEX order_status_idx ON "order" (status);
	`
	_, err := tx.ExecContext(ctx, query)
	return err
}
