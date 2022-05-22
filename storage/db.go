package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/alenapetraki/chat/storage/migrations"
	"github.com/pkg/errors"
)

type DB interface {
	Exec(string, ...interface{}) (sql.Result, error)
	Query(string, ...interface{}) (*sql.Rows, error)
	QueryRow(string, ...interface{}) *sql.Row
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
	Begin() (*Transaction, error)
	RunTx(fn func(tx *Transaction) error) error
}

func NewDB(sqldb *sql.DB) DB {
	return &db{db: sqldb}
}

type db struct {
	db *sql.DB
}

func (d *db) Exec(sql string, args ...interface{}) (sql.Result, error) {
	return d.db.Exec(sql, args...)
}

func (d *db) Query(sql string, args ...interface{}) (*sql.Rows, error) {
	return d.db.Query(sql, args...)
}

func (d *db) QueryRow(sql string, args ...interface{}) *sql.Row {
	return d.db.QueryRow(sql, args...)
}

func (d *db) ExecContext(ctx context.Context, sql string, args ...interface{}) (sql.Result, error) {
	return d.db.ExecContext(ctx, sql, args...)
}

func (d *db) QueryContext(ctx context.Context, sql string, args ...interface{}) (*sql.Rows, error) {
	return d.db.QueryContext(ctx, sql, args...)
}

func (d *db) QueryRowContext(ctx context.Context, sql string, args ...interface{}) *sql.Row {
	return d.db.QueryRowContext(ctx, sql, args...)
}

func (d *db) Begin() (*Transaction, error) {
	tx, err := d.db.BeginTx(context.Background(), nil)
	if err != nil {
		return nil, err
	}
	return &Transaction{tx: tx}, nil
}

func (d *db) RunTx(fn func(tx *Transaction) error) error {
	tx, err := d.Begin()
	if err != nil {
		return err
	}
	defer func() {
		p := recover()
		switch {
		case p != nil:
			// a panic occurred, rollback and repanic
			_ = tx.Rollback()
			panic(p)
		case err != nil:
			// something went wrong, rollback
			_ = tx.Rollback()
		default:
			// all good, commit
			err = tx.Commit()
		}
	}()
	err = fn(tx)
	return err
}

type Transaction struct {
	//DB
	tx *sql.Tx
}

func (t *Transaction) Exec(sql string, args ...interface{}) (sql.Result, error) {
	return t.tx.Exec(sql, args...)
}

func (t *Transaction) Query(sql string, args ...interface{}) (*sql.Rows, error) {
	return t.tx.Query(sql, args...)
}

func (t *Transaction) QueryRow(sql string, args ...interface{}) *sql.Row {
	return t.tx.QueryRow(sql, args...)
}

func (t *Transaction) ExecContext(ctx context.Context, sql string, args ...interface{}) (sql.Result, error) {
	return t.tx.ExecContext(ctx, sql, args...)
}

func (t *Transaction) QueryContext(ctx context.Context, sql string, args ...interface{}) (*sql.Rows, error) {
	return t.tx.QueryContext(ctx, sql, args...)
}

func (t *Transaction) QueryRowContext(ctx context.Context, sql string, args ...interface{}) *sql.Row {
	return t.tx.QueryRowContext(ctx, sql, args...)
}

func (t *Transaction) Commit() error {
	return t.tx.Commit()
}

func (t *Transaction) Rollback() error {
	return t.tx.Rollback()
}

func (t *Transaction) Begin() (*Transaction, error) {
	// интересная история с SAVEPOINT вместо возврата ошибки
	return nil, errors.New("cannot begin tx on tx")
}

// RunTx exec sql with transaction
func (t *Transaction) RunTx(_ func(tx *Transaction) error) error {
	return errors.New("cannot begin tx on tx")
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
