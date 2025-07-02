package migration

import (
	"github.com/eggnocent/app-eccomerce-backend/pkg/logging"
	"github.com/jmoiron/sqlx"
)

var (
	logger *logging.Logger
	opt    = Option{}
)

type Option struct {
	SchemaDir string
	SeedDir   string
	DBPool    *sqlx.DB
}

func Init(lg *logging.Logger, o Option) {
	opt = o
	logger = lg

	if err := Schema(o.DBPool); err != nil {
		logger.Err.Errorf("failed to migrate schema: %v", err)
	}
	if err := Seed(o.DBPool); err != nil {
		logger.Err.Errorf("failed to migrate seed: %v", err)
	}
}
