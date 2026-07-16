package database

import (
	"fmt"
	"net/url"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func EnsureDatabaseExists(databaseURL string) error {
	dbName, err := parseDBName(databaseURL)
	if err != nil {
		return err
	}

	u, _ := url.Parse(databaseURL)
	host := u.Hostname()
	port := u.Port()
	if port == "" {
		port = "5432"
	}

	password, _ := u.User.Password()
	user := u.User.Username()

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=postgres", host, port, user, password)

	bootstrapDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to postgres database for bootstrap: %w", err)
	}

	sqlDB, err := bootstrapDB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	defer sqlDB.Close()

	var exists int
	if err := bootstrapDB.Raw("SELECT COUNT(*) FROM pg_database WHERE datname = ?", dbName).Scan(&exists).Error; err != nil {
		return fmt.Errorf("failed to check database existence: %w", err)
	}
	if exists == 0 {
		quotedName := fmt.Sprintf(`"%s"`, dbName)
		if err := bootstrapDB.Exec(fmt.Sprintf("CREATE DATABASE %s", quotedName)).Error; err != nil {
			return fmt.Errorf("failed to create database %s: %w", dbName, err)
		}
	}

	return nil
}

func parseDBName(databaseURL string) (string, error) {
	u, err := url.Parse(databaseURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse DATABASE_URL: %w", err)
	}

	dbName := strings.TrimPrefix(u.Path, "/")
	if idx := strings.Index(dbName, "?"); idx != -1 {
		dbName = dbName[:idx]
	}
	if dbName == "" {
		return "", fmt.Errorf("could not extract database name from DATABASE_URL")
	}
	return dbName, nil
}
