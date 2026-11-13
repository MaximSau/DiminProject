package repository

import (
	"database/sql"
	"embed"
	"fmt"
	"log/slog"

	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

// RunMigrations накатывает встроенные SQL-миграции с помощью goose.
func RunMigrations(db *sql.DB) error {
	goose.SetBaseFS(embedMigrations)
	if err := goose.SetDialect("sqlite3"); err != nil {
		return fmt.Errorf("ошибка настройки диалекта goose sqlite3: %w", err)
	}

	if err := goose.Up(db, "migrations"); err != nil {
		return fmt.Errorf("ошибка применения миграций goose: %w", err)
	}

	slog.Info("база данных успешно актуализирована через goose миграции")
	return nil
}
