package files

import (
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigration(upInitialSchema, downInitialSchema)
}

func upInitialSchema(tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	return nil
}

func downInitialSchema(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	return nil
}
