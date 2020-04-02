package driver

import (
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/interactive-solutions/go-sql-migrations"
)

var _ migrations.Driver = (*postgresDriver)(nil)

type postgresDriver struct {
	database *sqlx.DB
}

func NewPostgresDriver(database *sqlx.DB) *postgresDriver {
	return &postgresDriver{database}
}

func (d *postgresDriver) CreateVersionsTable() error {
	_, err := d.database.Exec(`CREATE TABLE IF NOT EXISTS database_versions(version VARCHAR(255) UNIQUE NOT NULL);`)
	return err
}

func (d *postgresDriver) HasExecuted(version string) bool {
	var count int
	err := d.database.Get(&count, "SELECT COUNT(*) FROM database_versions WHERE version=$1", version)
	if err != nil {
		panic(err)
	}

	return count > 0
}

func (d *postgresDriver) Up(migration migrations.Migration) error {

	tx, err := d.database.Begin()
	if err != nil {
		panic(err)
	}

	if _, err := tx.Exec(migration.Content.Up); err != nil {
		tx.Rollback()

		return errors.Wrapf(err, "failed to execute migration: %s", migration.Name)
	}

	if _, err := tx.Exec("INSERT INTO database_versions (version) VALUES ($1)", migration.VersionAsString()); err != nil {
		tx.Rollback()

		return errors.Wrapf(err, "persist migration upgrade to: %s", migration.Name)

	}

	return errors.Wrapf(tx.Commit(), "failed to commit for: %s", migration.Name)
}

func (d *postgresDriver) Down(migration migrations.Migration) error {

	tx, err := d.database.Begin()
	if err != nil {
		panic(err)
	}

	if _, err := tx.Exec(migration.Content.Down); err != nil {
		tx.Rollback()

		return errors.Wrapf(err, "failed to execute migration: %s", migration.Name)
	}

	if _, err := tx.Exec("DELETE FROM database_versions WHERE version = ($1)", migration.VersionAsString()); err != nil {
		tx.Rollback()

		return errors.Wrapf(err, "persist migration downgrade to: %s", migration.Name)
	}

	return errors.Wrapf(tx.Commit(), "failed to commit for: %s", migration.Name)
}
