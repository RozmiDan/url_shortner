package redirect_handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/RozmiDan/url_shortener/internal/storage"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

type URLGetter interface {
	GetURL(alias string) (string, error)
}

type Response struct {
	Error  string `json:"error,omitempty"`
	Status string `json:"status"`
	URL    string `json:"url,omitempty"`
}

// @Title Get URL by alias
// @Description Return URL for redirect to original by short alias
// @Tags redirect
// @Accept  json
// @Produce json
// @Param   alias  path  string  true  "Short URL alias"
// @Success 200 {string} string "Redirect to original URL"
// @Failure 404 {object} Response "Alias not found"
// @Failure 500 {object} Response "Internal server error"
// @Router /{alias} [get]
func NewRedirectHandler(logger *slog.Logger, urlGetter URLGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		const op = "redirect_handler.RedirectHandlerConstruction"

		logger = logger.With(slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())))

		reqAlias := chi.URLParam(r, "alias")

		//logger.Info("request alias is valid")

		url, err := urlGetter.GetURL(reqAlias)

		if err != nil {
			if errors.Is(err, storage.ErrURLNotFound) {
				logger.Debug("URL not found", slog.String("alias", reqAlias))
				render.Status(r, http.StatusNotFound)
				render.JSON(w, r, Response{
					Status: "Error",
					Error:  "URL not found",
				})
				return
			}

			logger.Error("Error while getting url", slog.Any("err", err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, Response{
				Status: "Error",
				Error:  "internal error",
			})
			return
		}

		//logger.Info("url was found", slog.String("url", url))

		// http.Redirect(w, r, url, http.StatusFound)

		render.Status(r, http.StatusOK)
		render.JSON(w, r, Response{
			Status: "OK",
			URL:    url,
		})
	}
}
