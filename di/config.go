package di

import (
	"fkirill.org/eink-meteo-station/config"
	"github.com/google/wire"
)

func provideConfig() (config.ConfigApi, error) {
	return config.NewConfigApi()
}

var configModule = wire.NewSet(
	provideConfig,
)
