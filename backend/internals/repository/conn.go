package repository

import (
	"log"
	"strings"

	"github.com/jmoiron/sqlx"
	migrate "github.com/rubenv/sql-migrate"
)

// OpenDatabase Open up the database connection
func OpenDatabase(driver string, connection string, schema string) (*sqlx.DB, error) {
	conn := strings.ReplaceAll(connection, "{schema}", schema)
	db, err := sqlx.Open(driver, conn)
	if err != nil {
		log.Printf("Failed to open schema store")
		return nil, err
	}

	// Run database migrations
	err = updateStore(driver, db)
	if err != nil {
		log.Printf("Failed to upgrade the schema store")
		return nil, err
	}

	return db, nil
}

// updateStore Run any migrations that need to run
func updateStore(driver string, db *sqlx.DB) error {
	migrations := &migrate.FileMigrationSource{
		Dir: "migrations",
	}
	n, err := migrate.Exec(db.DB, driver, migrations, migrate.Up)
	if err != nil {
		log.Printf("Filed migrations\n")
		return err
	}
	log.Printf("Applied %d migrations\n", n)
	return nil
}
