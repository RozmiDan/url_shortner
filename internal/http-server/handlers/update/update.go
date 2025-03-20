package update_handler

import (
	"log/slog"
	"net/http"

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
	Alias  string `json:"alias"`
}

func NewUpdateHandler(logger *slog.Logger, storage URLUpdater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.update.newupdatehandler"

		logger = logger.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		reqAlias := chi.URLParam(r, "alias")

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			logger.Error("failed to decode request body")
			render.JSON(w, r, Response{
				Status: "Error",
				Error:  "failed to decode request",
			})
			return
		}

		logger.Info("request body decoded", slog.Any("request", req))

		if req.NewAlias == "" {
			alias = random.NewAliasForURL(aliasLength)
		}




	}
}
