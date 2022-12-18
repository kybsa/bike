package db

import (
	"errors"
	"testing"

	"github.com/kybsa/bike/config"
	"gorm.io/gorm"
)

func Test_GivenOpenReturnDB_WhenNewPostgresComponent_ThenReturnNilError(t *testing.T) {
	// Given
	sqlOpen = func(dialector gorm.Dialector, opt ...gorm.Option) (*gorm.DB, error) {
		return &gorm.DB{}, nil
	}
	simpleConfig := &config.SimpleConfig{
		MapConfig: map[string]string{"PostgresComponent.Dsn": "ca"},
	}
	// When
	postgres, err := NewPostgresComponent(simpleConfig)
	// Then
	if postgres == nil {
		t.Error("NewPostgresComponent must return component not null")
	}

	if err != nil {
		t.Errorf("NewPostgresComponent must return nil error. Error:%s", err.Error())
	}
}

func Test_GivenInvalidConfig_WhenNewPostgresComponent_ThenReturnError(t *testing.T) {
	// Given
	sqlOpen = func(dialector gorm.Dialector, opt ...gorm.Option) (*gorm.DB, error) {
		return &gorm.DB{}, nil
	}
	simpleConfig := &config.SimpleConfig{}
	// When
	postgres, err := NewPostgresComponent(simpleConfig)
	// Then
	if postgres != nil {
		t.Error("NewPostgresComponent must return component must be not null")
	}

	if err == nil {
		t.Errorf("NewPostgresComponent must return an error.")
	}
}

func Test_GivenOpenReturnError_WhenNewPostgresComponent_ThenReturnError(t *testing.T) {
	// Given
	sqlOpen = func(dialector gorm.Dialector, opt ...gorm.Option) (*gorm.DB, error) {
		return nil, errors.New("error")
	}
	simpleConfig := &config.SimpleConfig{
		MapConfig: map[string]string{"PostgresComponent.Dsn": "ca"},
	}
	// When
	postgres, err := NewPostgresComponent(simpleConfig)
	// Then
	if postgres != nil {
		t.Error("NewPostgresComponent must return component must be not null")
	}

	if err == nil {
		t.Errorf("NewPostgresComponent must return an error.")
	}
}
