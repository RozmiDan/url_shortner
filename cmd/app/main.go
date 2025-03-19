package main

import (
	"github.com/RozmiDan/url_shortener/internal/app"
	"github.com/RozmiDan/url_shortener/internal/config"
)

func main() {
	cnfg := config.MustLoad()

	app.Run(cnfg)
}
