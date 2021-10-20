package sunset_sunrise

import (
	"image"
	"math"
	"renderable"
	"renderable/puppettier"
	"renderable/utils"
	"strconv"
	"time"
)

type SunsetSunriseData struct {
	SunriseTime string // five characters
	SunsetTime  string //  five characters
}

type sunriseSunsetRenderable struct {
	offset             image.Point
	size               image.Point
	raster             []byte
	nextRedrawDateTime time.Time
	timeProvider       utils.TimeProvider
}

func (s *sunriseSunsetRenderable) RedrawNow() {
	s.nextRedrawDateTime = s.timeProvider.Now()
}

func NewSunriseSunsetRenderable(offset image.Point, timeProvider utils.TimeProvider) renderable.Renderable {
	size := image.Point{X: 420, Y: 400}
	raster := make([]byte, size.X*size.Y, size.X*size.Y)
	for i := range raster {
		raster[i] = 0xff
	}
	return &sunriseSunsetRenderable{
		offset:             offset,
		size:               size,
		raster:             raster,
		nextRedrawDateTime: timeProvider.Now(),
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

func (s *sunriseSunsetRenderable) NextRedrawDateTime() time.Time {
	return s.nextRedrawDateTime
}

func (s *sunriseSunsetRenderable) RedrawFinished() {
	s.nextRedrawDateTime = s.timeProvider.Now().Truncate(time.Hour * 24).Add(time.Hour * 24)
}

const picnicPointLatitude = -33.969526
const piclicPointLongitude = 150.998711

func (s *sunriseSunsetRenderable) Render() error {
	now := time.Now()
	_, tzOffset := now.Zone()
	timeOffsetInHours := float64(tzOffset) / 3600.0

	sunset := NewSunSet(picnicPointLatitude, piclicPointLongitude, timeOffsetInHours)
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

func (s *sunriseSunsetRenderable) DisplayMode() int {
	return 1
}

func (s *sunriseSunsetRenderable) String() string {
	return "sunrise_sunset"
}
