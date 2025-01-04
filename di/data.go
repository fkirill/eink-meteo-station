package di

import (
	"fkirill.org/eink-meteo-station/config"
	"fkirill.org/eink-meteo-station/data/daylight"
	"fkirill.org/eink-meteo-station/data/environment"
	"fkirill.org/eink-meteo-station/data/ha"
	"fkirill.org/eink-meteo-station/data/weather"
	"github.com/google/wire"
)

func provideForecastData(cfg config.ConfigApi) (weather.ForecastDataProvider, error) {
	return weather.NewForecastDataProvider(cfg)
}

func provideHomeAssistantApi(cfg config.ConfigApi) ha.HomeAssistantApi {
	return ha.NewHomeAssistantApi(cfg)
}

func provideEnvironmentData(cfg config.ConfigApi, haApi ha.HomeAssistantApi) environment.EnvironmentDataProvider {
	return environment.NewEnvironmentDataProvider(cfg, haApi)
}

func provideSunriseSunsetProvider() daylight.SunriseSunsetProvider {
	return daylight.NewSunriseSunsetProvider()
}

var dataModule = wire.NewSet(
	provideForecastData,
	provideHomeAssistantApi,
	provideEnvironmentData,
	provideSunriseSunsetProvider,
)
