package forecast

import (
	"bytes"
	"fkirill.org/eink-meteo-station/clib"
	"fkirill.org/eink-meteo-station/data/weather"
	"fkirill.org/eink-meteo-station/puppettier"
	"fkirill.org/eink-meteo-station/renderable"
	"fkirill.org/eink-meteo-station/renderable/utils"
	"github.com/rotisserie/eris"
	"html/template"
	"image"
	"strconv"
	"time"
)

type forecastRenderable struct {
	weather                weather.ForecastDataProvider
	offset                 image.Point
	size                   image.Point
	raster                 []byte
	nextRedrawDateTime     time.Time
	timeProvider           utils.TimeProvider
	forecastParsedTemplate *template.Template
}

func (f *forecastRenderable) RedrawNow() {
	f.nextRedrawDateTime = f.timeProvider.UtcNow()
}

type ForecastRenderable interface {
	renderable.Renderable
}

func NewForecastRenderable(rect image.Rectangle, timeProvider utils.TimeProvider, weather weather.ForecastDataProvider) (ForecastRenderable, error) {
	rasterSize := rect.Dx() * rect.Dy()
	raster := make([]byte, rasterSize, rasterSize)
	for i, _ := range raster {
		raster[i] = 0xff
	}
	tmpl, err := template.New("forecast").Parse(forecastTemplate)
	if err != nil {
		return nil, eris.Wrap(err, "Cannot parse template")
	}
	return &forecastRenderable{
		weather:                weather,
		offset:                 rect.Min,
		size:                   rect.Size(),
		raster:                 raster,
		nextRedrawDateTime:     timeProvider.UtcNow(),
		timeProvider:           timeProvider,
		forecastParsedTemplate: tmpl,
	}, nil
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
	forecastData, err := f.weather.GetWeatherData()
	if err != nil {
		return err
	}
	// was unable to read the data this time or there was an error, will retry next time
	if forecastData.Days == nil || len(forecastData.Days) == 0 {
		return nil
	}
	html, err := f.generateForecastHtml(forecastData)
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

func (f *forecastRenderable) generateForecastHtml(forecastData *weather.ForecastData) (string, error) {
	forecastTable := convertToTemplateFormat(forecastData.Days)
	buffer := bytes.Buffer{}
	err := f.forecastParsedTemplate.Execute(&buffer, forecastTable)
	if err != nil {
		return "", err
	}
	return string(buffer.Bytes()), nil
}

func convertToTemplateFormat(days []weather.ForecastDataDay) *forecastTable {
	res := make([]*dailyForecast, 0)
	for _, day := range days {
		daily := &dailyForecast{
			DayOfMonth:   strconv.Itoa(day.Date.Day()),
			DayOfWeek:    day.Date.Weekday().String()[:3],
			Month:        day.Date.Month().String()[:3],
			MinTemp:      strconv.Itoa(int(day.MinTemp)),
			MaxTemp:      strconv.Itoa(int(day.MaxTemp)),
			AmountOfRain: strconv.Itoa(int(day.ExpectedRainAmountMm)),
			AmountOfSnow: strconv.Itoa(int(day.ExpectedSnowAmountMm)),
			MaxWind:      strconv.Itoa(int(day.MaxWindKmh)),
			WeatherType:  0,
		}
		res = append(res, daily)
	}
	return &forecastTable{
		Days: res,
	}
}
