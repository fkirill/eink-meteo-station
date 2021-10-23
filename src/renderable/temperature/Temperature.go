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

func NewHATemperatureView(
	offset image.Point,
	timeProvider utils.TimeProvider,
	insideTemperatureSensorFn,
	outsideTemperatureSensorFn,
	insideHumiditySensorFn,
	outsideHumiditySensorFn func() string) renderable.Renderable {
	raster := make([]byte, temperatureWidgetSize.X*temperatureWidgetSize.Y, temperatureWidgetSize.X*temperatureWidgetSize.Y)
	for i := range raster {
		raster[i] = 0xff
	}
	return &temperatureView{
		size:           temperatureWidgetSize,
		offset:         offset,
		nextRedrawTime: timeProvider.UtcNow(),
		raster:         raster,
		inside:         ha.TemperatureHumidityData{},
		outside:        ha.TemperatureHumidityData{},
		timeProvider:   timeProvider,
		insideTemperatureSensorFn: insideTemperatureSensorFn,
		insideHumiditySensorFn: insideHumiditySensorFn,
		outsideTemperatureSensorFn: outsideTemperatureSensorFn,
		outsideHumiditySensorFn: outsideHumiditySensorFn,
	}
}

type temperatureView struct {
	size                       image.Point
	offset                     image.Point
	nextRedrawTime             time.Time
	raster                     []byte
	inside                     ha.TemperatureHumidityData
	outside                    ha.TemperatureHumidityData
	timeProvider               utils.TimeProvider
	insideTemperatureSensorFn  func() string
	outsideTemperatureSensorFn func() string
	insideHumiditySensorFn     func() string
	outsideHumiditySensorFn    func() string
}

func (t *temperatureView) RedrawNow() {
	t.nextRedrawTime = t.timeProvider.UtcNow()
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

func (t *temperatureView) NextRedrawDateTimeUtc() time.Time {
	return t.nextRedrawTime
}

func (t *temperatureView) RedrawFinished() {
	// refresh at random intervals 200 to 400 seconds (approximately every 5 minutes)
	t.nextRedrawTime = t.timeProvider.UtcNow().Add(time.Second * time.Duration(rand.Intn(200)+200))
}

func (t *temperatureView) Render() error {
	inside, err := ha.GetInsideTemperatureHumidity(t.insideTemperatureSensorFn(), t.insideHumiditySensorFn())
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
	outside, err := ha.GetOutsideTemperatureHumidity(t.outsideTemperatureSensorFn(), t.outsideHumiditySensorFn())
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
