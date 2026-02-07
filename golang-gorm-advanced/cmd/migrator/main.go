package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/rinkachi/golang-demos/golang-gorm-advanced/internal/config"
	"github.com/rinkachi/golang-demos/golang-gorm-advanced/internal/infra/db"
	"github.com/rinkachi/golang-demos/golang-gorm-advanced/internal/migrations"
	"gorm.io/gorm"
)

func main() {
	cmd := "up"
	if len(os.Args) > 1 {
		cmd = strings.ToLower(os.Args[1])
	}

	cfg := config.LoadMigrator()
	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db connect failed: %v", err)
	}

	if err := migrations.EnsureSchemaTable(database); err != nil {
		log.Fatalf("ensure schema table: %v", err)
	}

	switch cmd {
	case "up":
		if err := applyUp(database); err != nil {
			log.Fatalf("migrate up failed: %v", err)
		}
	case "down":
		if err := applyDown(database); err != nil {
			log.Fatalf("migrate down failed: %v", err)
		}
	case "reset":
		if err := reset(database); err != nil {
			log.Fatalf("reset failed: %v", err)
		}
	case "status":
		if err := status(database); err != nil {
			log.Fatalf("status failed: %v", err)
		}
	default:
		log.Fatalf("unknown command: %s (use up|down|reset|status)", cmd)
	}
}

func applyUp(db *gorm.DB) error {
	applied, err := migrations.Applied(db)
	if err != nil {
		return err
	}

	for _, m := range migrations.List() {
		if _, ok := applied[m.ID]; ok {
			continue
		}

		if err := db.Transaction(func(tx *gorm.DB) error {
			if err := m.Up(tx); err != nil {
				return err
			}
			return migrations.RecordApplied(tx, m.ID)
		}); err != nil {
			return err
		}

		log.Printf("applied %s", m.ID)
	}

	return nil
}

func applyDown(db *gorm.DB) error {
	last, err := migrations.LastApplied(db)
	if err != nil {
		return err
	}

	var target *migrations.Migration
	for _, m := range migrations.List() {
		if m.ID == last.ID {
			target = &m
			break
		}
	}
	if target == nil {
		return fmt.Errorf("migration not found: %s", last.ID)
	}

	if err := db.Transaction(func(tx *gorm.DB) error {
		if err := target.Down(tx); err != nil {
			return err
		}
		return migrations.RemoveApplied(tx, target.ID)
	}); err != nil {
		return err
	}

	log.Printf("rolled back %s", target.ID)
	return nil
}

func reset(db *gorm.DB) error {
	applied, err := migrations.Applied(db)
	if err != nil {
		return err
	}
	if len(applied) == 0 {
		return errors.New("no migrations applied")
	}

	ordered := migrations.List()
	for i := len(ordered) - 1; i >= 0; i-- {
		m := ordered[i]
		if _, ok := applied[m.ID]; !ok {
			continue
		}
		if err := db.Transaction(func(tx *gorm.DB) error {
			if err := m.Down(tx); err != nil {
				return err
			}
			return migrations.RemoveApplied(tx, m.ID)
		}); err != nil {
			return err
		}
		log.Printf("rolled back %s", m.ID)
	}

	return applyUp(db)
}

func status(db *gorm.DB) error {
	applied, err := migrations.Applied(db)
	if err != nil {
		return err
	}

	for _, m := range migrations.List() {
		state := "pending"
		if _, ok := applied[m.ID]; ok {
			state = "applied"
		}
		log.Printf("%s\t%s", state, m.ID)
	}
	return nil
}
