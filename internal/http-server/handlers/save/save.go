package save_handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/RozmiDan/url_shortener/internal/storage"
	"github.com/RozmiDan/url_shortener/internal/usecase/random"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator"
)

const aliasLength = 6

type URLSaver interface {
	SaveURL(urlToSave string, alias string) (int64, error)
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
		const op = "handlers.save.newsavehandler"

		logger = logger.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

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

		logger.Info("request body decoded", slog.Any("request", req))

		validate := validator.New()
		if err := validate.Struct(req); err != nil {
			logger.Error("validation error", slog.Any("error", err))
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

		_, err = urlSaver.SaveURL(req.URL, alias)
		if err != nil {
			if errors.Is(err, storage.ErrURLExists) {
				logger.Error("URL already exists", slog.String("URL", req.URL))
				render.Status(r, http.StatusConflict)
				render.JSON(w, r, Response{
					Status: "Error",
					Error:  "URL already exists",
				})
				return
			}

			logger.Error("failed to save URL", slog.String("error", err.Error()))

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, Response{
				Status: "Error",
				Error:  "internal error",
			})
			return
		}

		logger.Info("Request has been successfuly done")

		render.Status(r, http.StatusOK)
		render.JSON(w, r, Response{
			Status: "OK",
			Alias:  alias,
		})
	}
}
