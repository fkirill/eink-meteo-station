package forecast

import (
	"fkirill.org/eink-meteo-station/clib"
	"fkirill.org/eink-meteo-station/renderable"
	"fkirill.org/eink-meteo-station/renderable/puppettier"
	"fkirill.org/eink-meteo-station/renderable/utils"
	"image"
	"strconv"
	"time"
)

type forecastRenderable struct {
	offset             image.Point
	size               image.Point
	raster             []byte
	nextRedrawDateTime time.Time
	timeProvider       utils.TimeProvider
}

func (f *forecastRenderable) RedrawNow() {
	f.nextRedrawDateTime = f.timeProvider.UtcNow()
}

var forecastSize = image.Point{X: 871, Y: 500}

func NewForecastRenderable(offset image.Point, timeProvider utils.TimeProvider) renderable.Renderable {
	raster := make([]byte, forecastSize.X*forecastSize.Y, forecastSize.X*forecastSize.Y)
	for i, _ := range raster {
		raster[i] = 0xff
	}
	return &forecastRenderable{
		offset:             offset,
		size:               forecastSize,
		raster:             raster,
		nextRedrawDateTime: timeProvider.UtcNow(),
		timeProvider:       timeProvider,
	}
}

func (f *forecastRenderable) BoundingBox() image.Rectangle {
	return utils.BoundingBox(f.offset, f.size)
}

func (f *forecastRenderable) Offset() image.Point {
	return f.offset
}

func (f *forecastRenderable) Size() image.Point {
	return f.size
}

func (f *forecastRenderable) Raster() []byte {
	return f.raster
}

func (f *forecastRenderable) NextRedrawDateTimeUtc() time.Time {
	return f.nextRedrawDateTime
}

func (f *forecastRenderable) RedrawFinished() {
	// redraw every 3 hours
	f.nextRedrawDateTime = f.timeProvider.UtcNow().Truncate(time.Hour).Add(3 * time.Hour)
}

func (f *forecastRenderable) Render() error {
	forecastData, err := GetWeatherData()
	if err != nil {
		return err
	}
	// was unable to read the data this time or there was an error, will retry next time
	if forecastData.Days == nil || len(forecastData.Days) == 0 {
		return nil
	}
	html, err := GenerateForecastHtml(forecastData)
	if err != nil {
		return err
	}
	img, err := puppettier.RenderInPuppeteer(html, "pressure_"+strconv.FormatInt(time.Now().Unix(), 10), f.size)
	if err != nil {
		return err
	}
	f.raster = img
	return nil
}

func (f *forecastRenderable) DisplayMode() uint8 {
	return clib.A2_Mode
}

func (f *forecastRenderable) String() string {
	return "forecast"
}
