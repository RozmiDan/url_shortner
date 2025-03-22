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
	delete_handler "github.com/RozmiDan/url_shortener/internal/http-server/handlers/delete"
	redirect_handler "github.com/RozmiDan/url_shortener/internal/http-server/handlers/redirect"
	save_handler "github.com/RozmiDan/url_shortener/internal/http-server/handlers/save"
	update_handler "github.com/RozmiDan/url_shortener/internal/http-server/handlers/update"
	middleware_logger "github.com/RozmiDan/url_shortener/internal/http-server/middleware"
	"github.com/RozmiDan/url_shortener/internal/storage/postgre"
	"github.com/RozmiDan/url_shortener/pkg/logger"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
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

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware_logger.MyLogger(logger))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Post("/url", save_handler.NewSaveHandler(logger, storage))
	router.Get("/{alias}", redirect_handler.NewUpdateHandler(logger, storage))
	router.Put("/url/{alias}", update_handler.NewUpdateHandler(logger, storage))
	router.Delete("/url/{alias}", delete_handler.NewDeleteHandler(logger, storage))

	server := http.Server{
		Addr:         cnfg.HttpInfo.Port,
		Handler:      router,
		ReadTimeout:  cnfg.HttpInfo.Timeout,
		WriteTimeout: cnfg.HttpInfo.Timeout,
		IdleTimeout:  cnfg.HttpInfo.IdleTimeout,
	}

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
