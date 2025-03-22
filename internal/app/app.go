package app

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/RozmiDan/url_shortener/db"

	"github.com/RozmiDan/url_shortener/internal/config"
	"github.com/RozmiDan/url_shortener/internal/http-server/server"
	"github.com/RozmiDan/url_shortener/internal/storage/postgre"
	"github.com/RozmiDan/url_shortener/pkg/logger"
	"github.com/jackc/pgx"
)

func Run(cnfg *config.Config) {

	logger := logger.NewLogger(cnfg.Env)

	logger.Info("url-shortner started")
	logger.Debug("debug mode")

	pgxConf := pgx.ConnConfig{
		Host:     cnfg.PostgreURL.Host,
		Port:     cnfg.PostgreURL.Port,
		Database: cnfg.PostgreURL.Database,
		User:     cnfg.PostgreURL.User,
		Password: cnfg.PostgreURL.Password,
	}

	db.SetupPostgres(pgxConf, logger)
	logger.Info("Migrations completed successfully\n")

	storage, err := postgre.New(cnfg.PostgreURL.URL)
	if err != nil {
		logger.Error("Cant open database", slog.Any("err", err))
		os.Exit(1)
	}

	logger.Info("Connected postgres\n")

	server := server.InitServer(cnfg, logger, storage)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		logger.Info("starting server", slog.String("port", cnfg.HttpInfo.Port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server error", slog.Any("err", err))
			os.Exit(1)
		}
	}()

	<-stop
	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Завершаем работу сервера
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server shutdown error", slog.Any("err", err))
	} else {
		logger.Info("Server gracefully stopped")
	}

	logger.Info("Finishing programm")
}
