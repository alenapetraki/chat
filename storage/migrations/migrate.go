package migrations

import (
	"database/sql"
	"embed"

	"github.com/pressly/goose/v3"
)

//go:embed files/*.sql
var embedMigrations embed.FS

func Migrate(db *sql.DB) error {

	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	if err := goose.Up(db, "files"); err != nil {
		return err
	}

	return nil
}
