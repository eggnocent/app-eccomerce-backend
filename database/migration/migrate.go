package migration

import (
	"github.com/jmoiron/sqlx"
	migrate "github.com/rubenv/sql-migrate"
)

func Schema(db *sqlx.DB) error {
	migrations := &migrate.FileMigrationSource{
		Dir: opt.SchemaDir,
	}

	migrate.SetTable("gorp_schema")

	n, err := migrate.Exec(db.DB, "postgres", migrations, migrate.Up)
	if err != nil {
		return err
	}
	logger.Out.Infof("Schema successfully migrated: %v", n)

	return nil
}

func Seed(db *sqlx.DB) error {
	migrations := &migrate.FileMigrationSource{
		Dir: opt.SeedDir,
	}

	migrate.SetTable("gorp_seed")

	n, err := migrate.Exec(db.DB, "postgres", migrations, migrate.Up)
	if err != nil {
		return err
	}

	logger.Out.Infof("Seed successfully migrated: %v", n)

	return nil
}
