package delete_handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/RozmiDan/url_shortener/internal/storage"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

type URLDeleter interface {
	DeleteURL(alias string) error
}

type Response struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

func NewDeleteHandler(logger *slog.Logger, urlDeleter URLDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.delete.newsavehandler"

		logger = logger.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		reqAlias := chi.URLParam(r, "alias")

		if reqAlias == "" {
			logger.Error("empty current alias")
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, Response{
				Status: "Error",
				Error:  "empty current alias",
			})
			return
		}

		logger.Info("request body decoded\n")

		if err := urlDeleter.DeleteURL(reqAlias); err != nil {
			if errors.Is(err, storage.ErrAliasNotFound) {
				logger.Error("Cant delete alias\n", slog.Any("err", err))
				render.Status(r, http.StatusNotFound)
				render.JSON(w, r, Response{
					Status: "Error",
					Error:  "alias not found",
				})
				return
			}

			logger.Error("Cant delete alias\n", slog.Any("err", err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, Response{
				Status: "Error",
				Error:  "internal error",
			})
			return
		}

		logger.Info("Elias has been successfully deleted")

		render.Status(r, http.StatusOK)
		render.JSON(w, r, Response{
			Status: "OK",
		})
	}
}
