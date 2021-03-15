package forecast

import (
	"image"
	"renderable"
	"renderable/puppettier"
	"renderable/utils"
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

var forecastSize = image.Point{X: 920, Y: 500}

func NewForecastRenderable(offset image.Point, timeProvider utils.TimeProvider) renderable.Renderable {
	raster := make([]byte, forecastSize.X*forecastSize.Y, forecastSize.X*forecastSize.Y)
	for i, _ := range raster {
		raster[i] = 0xff
	}
	return &forecastRenderable{
		offset:             offset,
		size:               forecastSize,
		raster:             raster,
		nextRedrawDateTime: timeProvider.Now(),
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

func (f *forecastRenderable) NextRedrawDateTime() time.Time {
	return f.nextRedrawDateTime
}

func (f *forecastRenderable) RedrawFinished() {
	// redraw every 3 hours
	f.nextRedrawDateTime = f.timeProvider.Now().Truncate(time.Hour).Add(3 * time.Hour)
}

func (f *forecastRenderable) Render() error {
	forecastData, err := GetWeatherData()
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

func (f *forecastRenderable) DisplayMode() int {
	return 1
}

func (f *forecastRenderable) String() string {
	return "forecast"
}
