package config

import "os"

type APIConfig struct {
	HTTPAddr    string
	MetricsAddr string
	DatabaseURL string
}

type MigratorConfig struct {
	DatabaseURL string
}

func LoadAPI() APIConfig {
	return APIConfig{
		HTTPAddr:    getEnv("HTTP_ADDR", ":8086"),
		MetricsAddr: getEnv("METRICS_ADDR", ":9111"),
		DatabaseURL: getEnv("DATABASE_URL", "host=localhost user=user password=password dbname=gorm_advanced port=5432 sslmode=disable"),
	}
}

func LoadMigrator() MigratorConfig {
	return MigratorConfig{
		DatabaseURL: getEnv("DATABASE_URL", "host=localhost user=user password=password dbname=gorm_advanced port=5432 sslmode=disable"),
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
