package server

import (
	"context"
	"log/slog"
	"net/http"

	_ "github.com/RozmiDan/url_shortener/docs"
	"github.com/RozmiDan/url_shortener/internal/config"
	delete_handler "github.com/RozmiDan/url_shortener/internal/http-server/handlers/delete"
	redirect_handler "github.com/RozmiDan/url_shortener/internal/http-server/handlers/redirect"
	save_handler "github.com/RozmiDan/url_shortener/internal/http-server/handlers/save"
	update_handler "github.com/RozmiDan/url_shortener/internal/http-server/handlers/update"
	middleware_logger "github.com/RozmiDan/url_shortener/internal/http-server/middleware/logger"
	middleware_metrics "github.com/RozmiDan/url_shortener/internal/http-server/middleware/metrics"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"
)

type DataBase interface {
	SaveURL(ctx context.Context, urlToSave string, alias string) (int64, error)
	GetURL(alias string) (string, error)
	DeleteURL(alias string) error
	UpdateURL(currAlias string, newAlias string) error
}

func InitServer(cnfg *config.Config, logger *slog.Logger, db DataBase) *http.Server {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware_logger.MyLogger(logger))
	router.Use(middleware_metrics.MetricsMiddleware)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	router.Post("/url", save_handler.NewSaveHandler(logger, db))
	router.Get("/{alias}", redirect_handler.NewRedirectHandler(logger, db))
	router.Get("/swagger/*", httpSwagger.WrapHandler)
	router.Put("/url/{alias}", update_handler.NewUpdateHandler(logger, db))
	router.Delete("/url/{alias}", delete_handler.NewDeleteHandler(logger, db))
	router.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:         cnfg.HttpInfo.Port,
		Handler:      router,
		ReadTimeout:  cnfg.HttpInfo.Timeout,
		WriteTimeout: cnfg.HttpInfo.Timeout,
		IdleTimeout:  cnfg.HttpInfo.IdleTimeout,
	}

	return server
}
