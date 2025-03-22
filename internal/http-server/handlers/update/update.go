package update_handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/RozmiDan/url_shortener/internal/storage"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

type URLUpdater interface {
	UpdateURL(currAlias string, newAlias string) error
}

type Request struct {
	NewAlias string `json:"newAlias"`
}

type Response struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

// @Title Update URL alias
// @Description Update existing short URL alias
// @Tags url
// @Accept  json
// @Produce json
// @Param   alias  path  string  true  "Current short URL alias"
// @Param   input  body  Request  true  "New alias data"
// @Success 200 {object} Response
// @Failure 400 {object} Response "Invalid input data"
// @Failure 404 {object} Response "Alias not found"
// @Failure 409 {object} Response "New alias already exists"
// @Failure 500 {object} Response "Internal server error"
// @Router /url/{alias} [put]
func NewUpdateHandler(logger *slog.Logger, urlUpdater URLUpdater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.update.newupdatehandler"

		logger = logger.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		curAlias := chi.URLParam(r, "alias")
		if curAlias == "" {
			logger.Error("empty current alias")
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, Response{
				Status: "Error",
				Error:  "empty current alias",
			})
			return
		}

		var req Request
		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			logger.Error("failed to decode request body")
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, Response{
				Status: "Error",
				Error:  "failed to decode request",
			})
			return
		}

		logger.Info("request body decoded\n", slog.Any("request", req))

		newAlias := req.NewAlias

		if newAlias == curAlias {
			logger.Error("the same alias")
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, Response{
				Status: "Error",
				Error:  "new alias must be different",
			})
			return
		}

		if req.NewAlias == "" {
			logger.Error("new alias is empty")
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, Response{
				Status: "Error",
				Error:  "new alias is required",
			})
			return
		}

		if err := urlUpdater.UpdateURL(curAlias, newAlias); err != nil {
			if errors.Is(err, storage.ErrAliasExists) {
				logger.Error("Cant update alias\n", slog.Any("err", err))
				render.Status(r, http.StatusConflict)
				render.JSON(w, r, Response{
					Status: "Error",
					Error:  "alias already exists",
				})
				return
			} else if errors.Is(err, storage.ErrAliasNotFound) {
				logger.Error("Cant update alias\n", slog.Any("err", err))
				render.Status(r, http.StatusNotFound)
				render.JSON(w, r, Response{
					Status: "Error",
					Error:  "alias not found",
				})
				return
			}

			logger.Error("Cant update alias\n", slog.Any("err", err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, Response{
				Status: "Error",
				Error:  "internal error",
			})
			return
		}

		logger.Info("Elias has been successfully updated")

		render.Status(r, http.StatusOK)
		render.JSON(w, r, Response{
			Status: "OK",
		})
	}
}
