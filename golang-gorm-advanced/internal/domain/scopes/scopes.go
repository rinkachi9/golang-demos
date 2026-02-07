package scopes

import (
	"strings"
	"time"

	"gorm.io/gorm"
)

func ActiveUsers(db *gorm.DB) *gorm.DB {
	return db.Where("active = ?", true)
}

func EmailDomain(domain string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if strings.TrimSpace(domain) == "" {
			return db
		}
		return db.Where("email LIKE ?", "%@"+domain)
	}
}

func OrdersMinTotal(min float64) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if min <= 0 {
			return db
		}
		return db.Where("total >= ?", min)
	}
}

func OrdersRecentDays(days int) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if days <= 0 {
			return db
		}
		return db.Where("created_at >= ?", time.Now().AddDate(0, 0, -days))
	}
}
