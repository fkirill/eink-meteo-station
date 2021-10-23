package pressure

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

var pressureViewWidth = 400
var pressureViewHeight = 400
var pressureWidgetSize = image.Point{X: pressureViewWidth + 50, Y: pressureViewHeight}

func NewHAPressureView(offset image.Point, timeProvider utils.TimeProvider, pressureSensorFn func() string) renderable.Renderable {
	raster := make([]byte, pressureWidgetSize.X*pressureWidgetSize.Y, pressureWidgetSize.X*pressureWidgetSize.Y)
	for i := range raster {
		raster[i] = 0xff
	}
	return &pressureView{
		size:             pressureWidgetSize,
		offset:           offset,
		nextRedrawTime:   timeProvider.UtcNow(),
		raster:           raster,
		pressure:         ha.PressureData{},
		timeProvider:     timeProvider,
		pressureSensorFn: pressureSensorFn,
	}
}

type pressureView struct {
	size             image.Point
	offset           image.Point
	nextRedrawTime   time.Time
	raster           []byte
	pressure         ha.PressureData
	timeProvider     utils.TimeProvider
	pressureSensorFn func() string
}

func (p *pressureView) RedrawNow() {
	p.nextRedrawTime = p.timeProvider.UtcNow()
}

func (_ *pressureView) String() string {
	return "pressure"
}

func (p *pressureView) BoundingBox() image.Rectangle {
	return utils.BoundingBox(p.offset, p.size)
}

func (p *pressureView) Offset() image.Point {
	return p.offset
}

func (p *pressureView) Size() image.Point {
	return p.size
}

func (p *pressureView) Raster() []byte {
	return p.raster
}

func (p *pressureView) NextRedrawDateTimeUtc() time.Time {
	return p.nextRedrawTime
}

func (p *pressureView) RedrawFinished() {
	// refresh at random intervals 1800 to 2000 seconds (approximately every 5 minutes)
	p.nextRedrawTime = p.timeProvider.UtcNow().Add(time.Second * time.Duration(rand.Intn(200)+1800))
}

func (p *pressureView) Render() error {
	pressure, err := ha.GetPressure(p.pressureSensorFn())
	pressureNeedsRedraw := false
	if err != nil {
		if !p.pressure.Warning {
			p.pressure.Warning = true
			pressureNeedsRedraw = true
		}
	} else {
		if *pressure != p.pressure {
			p.pressure = *pressure
			pressureNeedsRedraw = true
		}
	}
	if pressureNeedsRedraw {
		html, err := GeneratePressureHtml(&p.pressure)
		if err != nil {
			return err
		}
		img, err := puppettier.RenderInPuppeteer(html, "pressure_"+strconv.FormatInt(time.Now().Unix(), 10), p.size)
		if err != nil {
			return err
		}
		p.raster = img
	}
	return nil
}

func (p *pressureView) DisplayMode() int {
	return 1
}
