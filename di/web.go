package di

import (
	"fkirill.org/eink-meteo-station/config"
	"fkirill.org/eink-meteo-station/webui"
	"github.com/google/wire"
)

func provideWebServer(cfg config.ConfigApi) webui.WebServer {
	return webui.NewWebServer(cfg)
}

var webModule = wire.NewSet(
	provideWebServer,
)
