package db

import (
	"embed"
	"log/slog"
	"os"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/stdlib"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

func SetupPostgres(conn pgx.ConnConfig, logger *slog.Logger) {
	goose.SetBaseFS(embedMigrations)
	if err := goose.SetDialect("postgres"); err != nil {
		logger.Error("can't set dialect in goose", slog.Any("err", err))
		os.Exit(1)
	}

	db := stdlib.OpenDB(conn)
	if err := goose.Up(db, "migrations"); err != nil {
		logger.Error("can't setup migrations", slog.Any("err", err))
		os.Exit(1)
	}
}
