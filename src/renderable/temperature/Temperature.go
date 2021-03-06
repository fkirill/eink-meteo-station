package temperature

import (
	"image"
	"math/rand"
	"renderable"
	"renderable/ha"
	"renderable/puppettier"
	"renderable/utils"
	"strconv"
	"time"
)

var tempViewWidth = 400
var tempViewHeight = 481
var signleTempViewSize = image.Point{X: tempViewWidth, Y: tempViewHeight}
var temperatureWidgetSize = image.Point{X: 2*tempViewWidth + 50, Y: tempViewHeight}

func NewHATemperatureView(offset image.Point, timeProvider utils.TimeProvider) renderable.Renderable {
	raster := make([]byte, temperatureWidgetSize.X*temperatureWidgetSize.Y, temperatureWidgetSize.X*temperatureWidgetSize.Y)
	for i := range raster {
		raster[i] = 0xff
	}
	return &temperatureView{
		size:           temperatureWidgetSize,
		offset:         offset,
		nextRedrawTime: timeProvider.Now(),
		raster:         raster,
		inside:         ha.TemperatureHumidityData{},
		outside:        ha.TemperatureHumidityData{},
		timeProvider:   timeProvider,
	}
}

type temperatureView struct {
	size           image.Point
	offset         image.Point
	nextRedrawTime time.Time
	raster         []byte
	inside         ha.TemperatureHumidityData
	outside        ha.TemperatureHumidityData
	timeProvider   utils.TimeProvider
}

func (_ *temperatureView) String() string {
	return "temperature"
}

func (t *temperatureView) BoundingBox() image.Rectangle {
	return utils.BoundingBox(t.offset, t.size)
}

func (t *temperatureView) Offset() image.Point {
	return t.offset
}

func (t *temperatureView) Size() image.Point {
	return t.size
}

func (t *temperatureView) Raster() []byte {
	return t.raster
}

func (t *temperatureView) NextRedrawDateTime() time.Time {
	return t.nextRedrawTime
}

func (t *temperatureView) RedrawFinished() {
	// refresh at random intervals 200 to 400 seconds (approximately every 5 minutes)
	t.nextRedrawTime = t.timeProvider.Now().Add(time.Second * time.Duration(rand.Intn(200)+200))
}

func (t *temperatureView) Render() error {
	inside, err := ha.GetInsideTemperatureHumidity()
	insideNeedsRedraw := false
	if err != nil {
		if !t.inside.Warning {
			t.inside.Warning = true
			insideNeedsRedraw = true
		}
	} else {
		if *inside != t.inside {
			t.inside = *inside
			insideNeedsRedraw = true
		}
	}
	outside, err := ha.GetOutsideTemperatureHumidity()
	outsideNeedsRedraw := false
	if err != nil {
		if !t.outside.Warning {
			t.outside.Warning = true
			outsideNeedsRedraw = true
		}
	} else {
		if *outside != t.outside {
			t.outside = *outside
			outsideNeedsRedraw = true
		}
	}
	if insideNeedsRedraw {
		html, err := GenerateTemperatureHtml(&t.inside)
		if err != nil {
			return err
		}
		img, err := puppettier.RenderInPuppeteer(html, "temperature_"+strconv.FormatInt(time.Now().Unix(), 10), signleTempViewSize)
		if err != nil {
			return err
		}
		utils.DrawImage(t.raster, t.size, image.Point{X: 0, Y: 0}, img, signleTempViewSize)
	}
	if outsideNeedsRedraw {
		html, err := GenerateTemperatureHtml(&t.outside)
		if err != nil {
			return err
		}
		img, err := puppettier.RenderInPuppeteer(html, "temperature_"+strconv.FormatInt(time.Now().Unix()+100, 10), signleTempViewSize)
		if err != nil {
			return err
		}
		utils.DrawImage(t.raster, t.size, image.Point{X: signleTempViewSize.X + 50, Y: 0}, img, signleTempViewSize)
	}
	return nil
}

func (t *temperatureView) DisplayMode() int {
	return 1
}
