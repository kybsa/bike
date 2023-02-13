// Package db contains data base features
package db

import (
	"errors"

	"github.com/kybsa/bike/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	sqlOpen = gorm.Open
)

// PostgresComponent to config postgres database client
type PostgresComponent struct {
	db *gorm.DB
}

func (postgresComponent *PostgresComponent) DB() *gorm.DB {
	return postgresComponent.db
}

func createDB(simpleConfig *config.SimpleConfig) (*gorm.DB, error) {
	dsn, ok := simpleConfig.Get("PostgresComponent.Dsn")
	if !ok {
		return nil, errors.New("error to get PostgresComponent.Dsn Config")
	}
	return sqlOpen(postgres.Open(dsn), &gorm.Config{})
}

func NewPostgresComponent(simpleConfig *config.SimpleConfig) (*PostgresComponent, error) {
	db, errDB := createDB(simpleConfig)
	if errDB != nil {
		return nil, errDB
	}
	return &PostgresComponent{
		db: db,
	}, nil
}

func NewPostgresComponentSession(
	simpleConfig *config.SimpleConfig,
	session *gorm.Session) (*PostgresComponent, error) {
	db, errDB := createDB(simpleConfig)
	if errDB != nil {
		return nil, errDB
	}
	return &PostgresComponent{
		db: db.Session(session),
	}, nil
}
