package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	Host         string
	Port         string
	User         string
	Password     string
	Database     string
	MinOpenConns int32
	MaxOpenConns int32
}

type Connection struct {
	config *pgxpool.Config
	pool   *pgxpool.Pool
}

func NewConnection(cfg Config) (*Connection, error) {
	source := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database)
	config, err := pgxpool.ParseConfig(source)
	if err != nil {
		return nil, err
	}

	config.MinConns = cfg.MinOpenConns
	config.MaxConns = cfg.MaxOpenConns

	return &Connection{
		config: config,
	}, nil
}

func (c *Connection) NewPool(ctx context.Context) error {
	var err error
	c.pool, err = pgxpool.NewWithConfig(ctx, c.config)
	if err != nil {
		return err
	}

	return c.pool.Ping(ctx)
}

func (c *Connection) GetPool() *pgxpool.Pool {
	return c.pool
}

func (c *Connection) ConnString() string {
	return c.pool.Config().ConnConfig.ConnString()
}

func (c *Connection) Close() {
	c.pool.Close()
}

func (c *Connection) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return c.pool.Query(ctx, sql, args...)
}

func (c *Connection) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return c.pool.QueryRow(ctx, sql, args...)
}

func (c *Connection) Exec(ctx context.Context, sql string, args ...any) error {
	_, err := c.pool.Exec(ctx, sql, args...)
	return err
}

func (c *Connection) Begin(ctx context.Context) (pgx.Tx, error) {
	return c.pool.Begin(ctx)
}
