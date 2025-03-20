package update_handler

import "log/slog"

type URLUpdater interface {
	UpdateURL(currAlias string, newAlias string) error
}

type Request struct {
	currAlias string	`json:"url" validate:"required,url"`
	newAlias string		`json:"url" validate:"required,url"`
}

type Response struct {

}

func NewUpdateHandler(logger *slog.Logger, storage URLUpdater) {

}
