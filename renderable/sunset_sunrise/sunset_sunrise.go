package sunset_sunrise

import (
	"bytes"
	"fkirill.org/eink-meteo-station/clib"
	"fkirill.org/eink-meteo-station/config"
	"fkirill.org/eink-meteo-station/data/daylight"
	"fkirill.org/eink-meteo-station/images"
	"fkirill.org/eink-meteo-station/puppettier"
	"fkirill.org/eink-meteo-station/renderable"
	"fkirill.org/eink-meteo-station/renderable/utils"
	"github.com/rotisserie/eris"
	"image"
	"strconv"
	"text/template"
	"time"
)

type SunsetSunriseData struct {
	SunriseTime string // five characters
	SunsetTime  string //  five characters
	SunrisePng  string
	SunsetPng   string
}

type sunriseSunsetRenderable struct {
	daylightProvider            daylight.SunriseSunsetProvider
	cfg                         config.ConfigApi
	offset                      image.Point
	size                        image.Point
	raster                      []byte
	nextRedrawDateTime          time.Time
	timeProvider                utils.TimeProvider
	sunsetSunriseParsedTemplate *template.Template
}

func (s *sunriseSunsetRenderable) RedrawNow() {
	s.nextRedrawDateTime = s.timeProvider.UtcNow()
}

type DaylightRenderable interface {
	renderable.Renderable
}

func NewSunriseSunsetRenderable(
	rect image.Rectangle,
	timeProvider utils.TimeProvider,
	cfg config.ConfigApi,
	daylightProvider daylight.SunriseSunsetProvider,
) (DaylightRenderable, error) {
	rasterSize := rect.Dx() * rect.Dy()
	raster := make([]byte, rasterSize, rasterSize)
	for i := range raster {
		raster[i] = 0xff
	}
	tmpl, err := template.New("sunsetSunrise").Parse(sunsetSunriseTemplate)
	if err != nil {
		return nil, eris.Wrap(err, "Cannot parse template")
	}
	return &sunriseSunsetRenderable{
		sunsetSunriseParsedTemplate: tmpl,
		daylightProvider:            daylightProvider,
		cfg:                         cfg,
		offset:                      rect.Min,
		size:                        rect.Size(),
		raster:                      raster,
		nextRedrawDateTime:          timeProvider.UtcNow(),
		timeProvider:                timeProvider,
	}, nil
}

func (s *sunriseSunsetRenderable) generateSunriseHtml(sunsetSunriseData *SunsetSunriseData) (string, error) {
	buffer := bytes.Buffer{}
	err := s.sunsetSunriseParsedTemplate.Execute(&buffer, sunsetSunriseData)
	if err != nil {
		return "", err
	}
	return string(buffer.Bytes()), nil
}

func (s *sunriseSunsetRenderable) BoundingBox() image.Rectangle {
	return utils.BoundingBox(s.offset, s.size)
}

func (s *sunriseSunsetRenderable) Offset() image.Point {
	return s.offset
}

func (s *sunriseSunsetRenderable) Size() image.Point {
	return s.size
}

func (s *sunriseSunsetRenderable) Raster() []byte {
	return s.raster
}

func (s *sunriseSunsetRenderable) NextRedrawDateTimeUtc() time.Time {
	return s.nextRedrawDateTime
}

func (s *sunriseSunsetRenderable) RedrawFinished() {
	s.nextRedrawDateTime = s.timeProvider.LocalNow().Truncate(time.Hour*24).AddDate(0, 0, 1).UTC()
}

func (s *sunriseSunsetRenderable) Render() error {
	latitude, longitude := s.cfg.GetDaylightCoordinates()
	sunrise, sunset := s.daylightProvider.GetSunriseSunset(latitude, longitude, s.timeProvider.LocalNow())

	sunriseSunsetData := SunsetSunriseData{
		SunriseTime: formatTime(sunrise),
		SunsetTime:  formatTime(sunset),
		SunrisePng:  images.Sunrise_png_src,
		SunsetPng:   images.Sunset_png_src,
	}
	html, err := s.generateSunriseHtml(&sunriseSunsetData)
	if err != nil {
		return err
	}
	raster, err := puppettier.RenderInPuppeteer(html, "sunrise_sunset", s.size)
	if err != nil {
		return err
	}
	s.raster = raster
	return nil
}

func formatTime(timePastMidnight time.Duration) string {
	hours := int(timePastMidnight.Hours())
	minutes := int(timePastMidnight.Minutes()) % 60
	hoursStr := strconv.Itoa(hours)
	if len(hoursStr) == 1 {
		hoursStr = "0" + hoursStr
	}
	minutesStr := strconv.Itoa(minutes)
	if len(minutesStr) == 1 {
		minutesStr = "0" + minutesStr
	}
	return hoursStr + ":" + minutesStr
}

func (s *sunriseSunsetRenderable) DisplayMode() uint8 {
	return clib.A2_Mode
}

func (s *sunriseSunsetRenderable) String() string {
	return "sunrise_sunset"
}
