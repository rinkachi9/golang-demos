package persistence

import (
	"fmt"
	"time"

	"github.com/rinkachi/golang-demos/golang-otel-demo/internal/domain"
	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewDB(dsn string) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	for i := 0; i < 15; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	if err := db.Use(otelgorm.NewPlugin()); err != nil {
		return nil, err
	}

	// AutoMigrate
	if err := db.AutoMigrate(&domain.ProcessLog{}); err != nil {
		return nil, err
	}

	return db, nil
}
