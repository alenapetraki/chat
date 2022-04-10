package postgres

import (
	"database/sql"
	"fmt"

	"github.com/pkg/errors"
)

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
}

func Connect(cfg *Config) (*sql.DB, error) {

	if cfg.Host == "" || cfg.Port == "" || cfg.User == "" ||
		cfg.Password == "" || cfg.Database == "" {
		return nil, errors.New("host:port, user:password and database parameters required")
	}

	db, err := sql.Open("postgres", fmt.Sprintf(
		"user=%s password=%s dbname=%s host=%s port=%s sslmode=disable",
		cfg.User, cfg.Password, cfg.Database, cfg.Host, cfg.Port))
	if err != nil {
		return nil, errors.Wrap(err, "failed to open connection to postgres database")
	}

	if err = db.Ping(); err != nil {
		return nil, errors.Wrap(err, "failed to ping")
	}

	return db, nil
}
