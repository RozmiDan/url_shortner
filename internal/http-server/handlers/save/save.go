package save_handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/RozmiDan/url_shortener/internal/storage"
	"github.com/RozmiDan/url_shortener/internal/usecase/random"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator"
)

const (
	aliasLength    = 6
	requestTimeout = 2 * time.Second
)

type URLSaver interface {
	SaveURL(ctx context.Context, urlToSave string, alias string) (int64, error)
}

type Request struct {
	URL   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty"`
}

type Response struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
	Alias  string `json:"alias,omitempty"`
}

// SaveURLHandler godoc
// @Summary      Creates a short URL
// @Description  Creates a short URL. If alias is not specified, a random string of 6 characters is generated.
// @Tags         url
// @Accept       json
// @Produce      json
// @Param        Request  body     Request  true  "URL Saving Parameters"
// @Success      200      {object} Response
// @Failure      400      {object} Response "invalid request parameters"
// @Failure      409      {object} Response "URL already exists"
// @Failure      500      {object} Response "Internal server error"
// @Router       /url [post]
func NewSaveHandler(logger *slog.Logger, urlSaver URLSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		opLogger := logger.With(
			slog.String("op", "handlers.save.NewSaveHandler"),
			slog.String("req_id", middleware.GetReqID(r.Context())),
		)

		var req Request
		if err := render.DecodeJSON(r.Body, &req); err != nil {
			opLogger.Debug("failed to decode request body", slog.Any("err", err))
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, Response{
				Status: "Error",
				Error:  "failed to decode request",
			})
			return
		}

		if err := validator.New().Struct(req); err != nil {
			opLogger.Debug("validation error", slog.Any("err", err))
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, Response{
				Status: "Error",
				Error:  "invalid request parameters",
			})
			return
		}

		alias := req.Alias
		if alias == "" {
			alias = random.NewAliasForURL(aliasLength)
		}

		ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
		defer cancel()

		_, err := urlSaver.SaveURL(ctx, req.URL, alias)
		if err != nil {
			if errors.Is(err, storage.ErrAliasExists) {
				opLogger.Debug("alias already exists", slog.String("alias", alias))
				render.Status(r, http.StatusConflict)
				render.JSON(w, r, Response{
					Status: "Error",
					Error:  "Alias already exists",
				})
				return
			}

			opLogger.Error("failed to save URL", slog.Any("err", err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, Response{
				Status: "Error",
				Error:  "internal error",
			})
			return
		}

		render.Status(r, http.StatusCreated)
		render.JSON(w, r, Response{
			Status: "OK",
			Alias:  alias,
		})
	}
}
