package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/alenapetraki/chat/storage/migrations"
	"github.com/pkg/errors"
)

type Conn interface {
	Exec(string, ...interface{}) (sql.Result, error)
	Query(string, ...interface{}) (*sql.Rows, error)
	QueryRow(string, ...interface{}) *sql.Row
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
	BeginTx(context.Context, *sql.TxOptions) (*sql.Tx, error)
}

type Tx interface {
	Commit() error
	Exec(string, ...interface{}) (sql.Result, error)
	Query(string, ...interface{}) (*sql.Rows, error)
	QueryRow(string, ...interface{}) *sql.Row
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
	Rollback() error
}

// tx - адаптер для использования как Conn
type transactioner struct {
	Tx
}

func (t *transactioner) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	panic("must not be called")
}

func NewTransactioner(tx Tx) Conn {
	return &transactioner{Tx: tx}
}

func RunTx(ctx context.Context, db *sql.DB, f func(tx *sql.Tx) error) error {

	var err error

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		p := recover()
		switch {
		case p != nil:
			_ = tx.Rollback()
			panic(p)
		case err != nil:
			_ = tx.Rollback()
		default:
			err = tx.Commit()
		}
	}()

	return f(tx)
}

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
}

func Connect(driver string, config *Config) (*sql.DB, error) {

	if config.Host == "" || config.Port == "" || config.User == "" ||
		config.Password == "" || config.Database == "" {
		return nil, errors.New("host:port, user:password and database parameters required")
	}

	db, err := sql.Open(driver, fmt.Sprintf(
		"user=%s password=%s dbname=%s host=%s port=%s sslmode=disable",
		config.User, config.Password, config.Database, config.Host, config.Port),
	)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	if err := setupDatabase(db, driver); err != nil {
		return nil, err
	}

	return db, nil
}

func setupDatabase(db *sql.DB, driver string) error {
	switch driver {
	case "postgres":
		return migrations.Migrate(db)
	default:
		return errors.Errorf("unknown db driver '%s'", driver)
	}
}
