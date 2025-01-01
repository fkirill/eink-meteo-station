package sunset_sunrise

import (
	"fkirill.org/eink-meteo-station/clib"
	"fkirill.org/eink-meteo-station/renderable"
	"fkirill.org/eink-meteo-station/renderable/puppettier"
	"fkirill.org/eink-meteo-station/renderable/utils"
	"image"
	"math"
	"strconv"
	"time"
)

type SunsetSunriseData struct {
	SunriseTime string // five characters
	SunsetTime  string //  five characters
}

type sunriseSunsetRenderable struct {
	latitude, longitude float64
	offset              image.Point
	size                image.Point
	raster              []byte
	nextRedrawDateTime  time.Time
	timeProvider        utils.TimeProvider
}

func (s *sunriseSunsetRenderable) RedrawNow() {
	s.nextRedrawDateTime = s.timeProvider.UtcNow()
}

func NewSunriseSunsetRenderable(offset image.Point, timeProvider utils.TimeProvider, latitude, longitude float64) renderable.Renderable {
	size := image.Point{X: 420, Y: 400}
	raster := make([]byte, size.X*size.Y, size.X*size.Y)
	for i := range raster {
		raster[i] = 0xff
	}
	return &sunriseSunsetRenderable{
		latitude:           latitude,
		longitude:          longitude,
		offset:             offset,
		size:               size,
		raster:             raster,
		nextRedrawDateTime: timeProvider.UtcNow(),
		timeProvider:       timeProvider,
	}
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
	now := time.Now()
	_, tzOffset := now.Zone()
	timeOffsetInHours := float64(tzOffset) / 3600.0

	//sunset := NewSunSet(picnicPointLatitude, picnicPointLongitude, timeOffsetInHours)
	sunset := NewSunSet(s.latitude, s.longitude, timeOffsetInHours)
	sunset.SetCurrentDate(s.nextRedrawDateTime.Year(), int(s.nextRedrawDateTime.Month()), s.nextRedrawDateTime.Day())
	sunriseTimeMinutes := sunset.CalcSunrise()
	sunsetTimeMinutes := sunset.CalcSunset()

	sunriseSunsetData := SunsetSunriseData{
		SunriseTime: fomatTimeInMinutes(sunriseTimeMinutes),
		SunsetTime:  fomatTimeInMinutes(sunsetTimeMinutes),
	}
	html, err := GenerateSunriseHtml(&sunriseSunsetData)
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

func fomatTimeInMinutes(minutesOfDay float64) string {
	hours := int(math.Trunc(minutesOfDay / 60.0))
	minutes := int(math.Trunc(minutesOfDay - 60.0*float64(hours)))
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
