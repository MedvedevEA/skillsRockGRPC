// Утилита миграция базы данных
// source-path ./../../migration/
// database-url postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable
package main

import (
	"flag"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	var (
		sourcePath  string
		databaseUrl string
	)
	flag.StringVar(&sourcePath, "source-path", "", "source path")
	flag.StringVar(&databaseUrl, "database-url", "", "database url")
	flag.Parse()
	if sourcePath == "" {
		log.Fatal("migration failed: source-path is required")
	}
	if databaseUrl == "" {
		log.Fatal("migration failed: database-url is required")
	}
	m, err := migrate.New(
		"file://"+sourcePath,
		databaseUrl,
	)
	if err != nil {
		log.Fatalf("migration failed: %v", err)
	}
	if err := m.Up(); err != nil {
		log.Fatalf("migration failed: %v", err)
	}
	log.Println("data migration successfully")
}
