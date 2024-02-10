package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(Up00001, nil)
}

func Up00001(ctx context.Context, tx *sql.Tx) error {
	query := `
		CREATE TABLE apikey (
			"token" varchar NOT NULL,
			service varchar NOT NULL,
			account_id varchar NOT NULL,
			account_uid uuid NOT NULL,
			create_ts int8 NOT NULL,
			CONSTRAINT apikey_pk PRIMARY KEY (token)
		);

		CREATE TABLE wallet (
			exchange varchar NOT NULL,
			account_uid uuid NOT NULL,
			coin varchar NOT NULL,
			total numeric(16, 8) NOT NULL DEFAULT 0,
			"hold" numeric(16, 8) NULL DEFAULT 0,
			CONSTRAINT wallet_unique UNIQUE (exchange, account_uid, coin)
		);

		CREATE TABLE "order" (
			order_uid uuid NOT NULL,
			account_uid uuid NOT NULL,
			exchange varchar NOT NULL,
			symbol varchar NOT NULL,
			"type" varchar NOT NULL,
			side varchar NOT NULL,
			amount numeric(16, 8) NOT NULL,
			price numeric(16, 8) NULL DEFAULT 0,
			hold_amount numeric(16, 8) NULL DEFAULT 0,
			hold_coin varchar NULL DEFAULT ''::character varying,
			status varchar NOT NULL,
			error varchar NULL,
			create_ts int8 NOT NULL,
			update_ts int8 NOT NULL,
			CONSTRAINT order_pk PRIMARY KEY (order_uid)
		);

		CREATE TABLE "position" (
			position_uid uuid NOT NULL,
			exchange varchar NOT NULL,
			account_uid uuid NOT NULL,
			symbol varchar NOT NULL,
			position_mode varchar NOT NULL,
			leverage int2 NULL,
			side varchar NOT NULL,
			amount numeric(16, 8) NOT NULL,
			price numeric(16, 8) NOT NULL,
			hold_amount numeric(16, 8) NULL DEFAULT 0,
			hold_coin varchar NULL DEFAULT ''::character varying,
			status varchar NOT NULL,
			create_ts int8 NOT NULL,
			update_ts int8 NOT NULL,
			CONSTRAINT position_pk PRIMARY KEY (position_uid)
		);
	`
	_, err := tx.ExecContext(ctx, query)
	return err
}
