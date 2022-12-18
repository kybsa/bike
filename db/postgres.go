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
	DB *gorm.DB
}

func NewPostgresComponent(simpleConfig *config.SimpleConfig) (*PostgresComponent, error) {
	dsn, ok := simpleConfig.Get("PostgresComponent.Dsn")
	if !ok {
		return nil, errors.New("error to get PostgresComponent.Dsn Config")
	}
	db, errDb := sqlOpen(postgres.Open(dsn), &gorm.Config{})
	if errDb != nil {
		return nil, errDb
	}
	return &PostgresComponent{
		DB: db,
	}, nil
}
