package migrations

import (
	"errors"
	"time"

	"github.com/rinkachi/golang-demos/golang-gorm-advanced/internal/domain/model"
	"gorm.io/gorm"
)

type Migration struct {
	ID   string
	Up   func(*gorm.DB) error
	Down func(*gorm.DB) error
}

type SchemaMigration struct {
	ID        string    `gorm:"primaryKey;size:64"`
	AppliedAt time.Time `gorm:"autoCreateTime"`
}

func List() []Migration {
	return []Migration{
		{
			ID: "001_create_tables",
			Up: func(db *gorm.DB) error {
				return db.AutoMigrate(
					&model.User{},
					&model.Order{},
					&model.OrderItem{},
					&model.AuditLog{},
				)
			},
			Down: func(db *gorm.DB) error {
				return db.Migrator().DropTable(
					&model.OrderItem{},
					&model.Order{},
					&model.AuditLog{},
					&model.User{},
				)
			},
		},
	}
}

func EnsureSchemaTable(db *gorm.DB) error {
	return db.AutoMigrate(&SchemaMigration{})
}

func Applied(db *gorm.DB) (map[string]SchemaMigration, error) {
	var rows []SchemaMigration
	if err := db.Find(&rows).Error; err != nil {
		return nil, err
	}
	applied := make(map[string]SchemaMigration, len(rows))
	for _, row := range rows {
		applied[row.ID] = row
	}
	return applied, nil
}

func RecordApplied(db *gorm.DB, id string) error {
	return db.Create(&SchemaMigration{ID: id}).Error
}

func RemoveApplied(db *gorm.DB, id string) error {
	return db.Delete(&SchemaMigration{ID: id}).Error
}

func LastApplied(db *gorm.DB) (SchemaMigration, error) {
	var row SchemaMigration
	if err := db.Order("applied_at desc, id desc").First(&row).Error; err != nil {
		return row, err
	}
	return row, nil
}

var ErrNoMigrations = errors.New("no migrations to apply")
