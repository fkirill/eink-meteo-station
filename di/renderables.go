package di

import (
	"fkirill.org/eink-meteo-station/config"
	"fkirill.org/eink-meteo-station/data/daylight"
	"fkirill.org/eink-meteo-station/data/environment"
	"fkirill.org/eink-meteo-station/data/weather"
	"fkirill.org/eink-meteo-station/eink"
	"fkirill.org/eink-meteo-station/renderable"
	"fkirill.org/eink-meteo-station/renderable/calendar"
	"fkirill.org/eink-meteo-station/renderable/clock"
	"fkirill.org/eink-meteo-station/renderable/forecast"
	"fkirill.org/eink-meteo-station/renderable/pressure"
	"fkirill.org/eink-meteo-station/renderable/sunset_sunrise"
	"fkirill.org/eink-meteo-station/renderable/temperature"
	"fkirill.org/eink-meteo-station/renderable/utils"
	"github.com/google/wire"
	"github.com/rotisserie/eris"
	"image"
)

type ScreenLayout struct {
	ScreenRect            image.Rectangle
	PressureWidgetRect    image.Rectangle
	CalendarWidgetRect    image.Rectangle
	ForecastWidgetRect    image.Rectangle
	ClockWidgetRect       image.Rectangle
	DaylightWidgetRect    image.Rectangle
	TemperatureWidgetRect image.Rectangle
}

// it doesn't belong here
func provideScreenLayout(eink eink.EInkScreen) (*ScreenLayout, error) {
	w, h := eink.GetScreenDimensions()
	if w != 1872 || h != 1404 {
		return nil, eris.Errorf("Unexpected screen dimentions: (%v, %v), expected (1872, 1404)", w, h)
	}
	return &ScreenLayout{
		ScreenRect:            image.Rectangle{Min: image.Point{X: 0, Y: 0}, Max: image.Point{X: int(w), Y: int(h)}},
		PressureWidgetRect:    image.Rectangle{Min: image.Point{X: 1000, Y: 500}, Max: image.Point{X: 1450, Y: 900}},
		CalendarWidgetRect:    image.Rectangle{Min: image.Point{X: 0, Y: 280}, Max: image.Point{X: 962, Y: 1400}},
		ForecastWidgetRect:    image.Rectangle{Min: image.Point{X: 1000, Y: 900}, Max: image.Point{X: 1871, Y: 1400}},
		ClockWidgetRect:       image.Rectangle{Min: image.Point{X: 0, Y: 0}, Max: image.Point{X: 963, Y: 237}},
		DaylightWidgetRect:    image.Rectangle{Min: image.Point{X: 1450, Y: 500}, Max: image.Point{X: 1870, Y: 900}},
		TemperatureWidgetRect: image.Rectangle{Min: image.Point{X: 1000, Y: 0}, Max: image.Point{X: 1850, Y: 481}},
	}, nil
}

func providePressureRenderable(
	layout *ScreenLayout,
	timeProvider utils.TimeProvider,
	cfg config.ConfigApi,
	envData environment.EnvironmentDataProvider,
) pressure.PressureRenderable {
	return pressure.NewHAPressureView(layout.PressureWidgetRect, timeProvider, cfg, envData)
}

func provideCalendarRenderable(
	layout *ScreenLayout,
	timeProvider utils.TimeProvider,
	cfg config.ConfigApi,
) calendar.CalendarRenderable {
	return calendar.NewCalendarRenderable(layout.CalendarWidgetRect, timeProvider, cfg)
}

func provideMultiRenderable(
	layout *ScreenLayout,
	pressureRenderable pressure.PressureRenderable,
	calendarWidget calendar.CalendarRenderable,
	forecastWidget forecast.ForecastRenderable,
	daylightWidget sunset_sunrise.DaylightRenderable,
	temperatureWidget temperature.TemperatureHumidityRenderable,
	clockWidget clock.ClockRenderable,
) (utils.MultiRenderable, error) {
	widgets := []renderable.Renderable{pressureRenderable, calendarWidget, forecastWidget, daylightWidget, temperatureWidget, clockWidget}
	res, err := utils.NewMultiRenderable(layout.ScreenRect, widgets, false)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func provideForecastRenderable(
	layout *ScreenLayout,
	timeProvider utils.TimeProvider,
	weather weather.ForecastDataProvider,
) (forecast.ForecastRenderable, error) {
	return forecast.NewForecastRenderable(layout.ForecastWidgetRect, timeProvider, weather)
}

func provideClockRenderable(
	layout *ScreenLayout,
	timeProvider utils.TimeProvider,
) (clock.ClockRenderable, error) {
	res, err := clock.NewClockRenderable(layout.ClockWidgetRect, timeProvider)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func provideDaylightRenderable(
	layout *ScreenLayout,
	timeProvider utils.TimeProvider,
	cfg config.ConfigApi,
	daylightProvider daylight.SunriseSunsetProvider,
) (sunset_sunrise.DaylightRenderable, error) {
	res, err := sunset_sunrise.NewSunriseSunsetRenderable(layout.DaylightWidgetRect, timeProvider, cfg, daylightProvider)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func provideTemperatureHumidityRenderable(
	layout *ScreenLayout,
	timeProvider utils.TimeProvider,
	envProvider environment.EnvironmentDataProvider,
) (temperature.TemperatureHumidityRenderable, error) {
	res, err := temperature.NewHATemperatureView(layout.TemperatureWidgetRect, timeProvider, envProvider)
	if err != nil {
		return nil, err
	}
	return res, nil
}

var renderableModule = wire.NewSet(
	providePressureRenderable,
	provideCalendarRenderable,
	provideMultiRenderable,
	provideForecastRenderable,
	provideClockRenderable,
	provideDaylightRenderable,
	provideTemperatureHumidityRenderable,
	provideScreenLayout,
)
