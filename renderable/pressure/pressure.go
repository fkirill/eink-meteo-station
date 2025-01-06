package pressure

import (
	"bytes"
	"fkirill.org/eink-meteo-station/clib"
	"fkirill.org/eink-meteo-station/config"
	"fkirill.org/eink-meteo-station/data/environment"
	"fkirill.org/eink-meteo-station/puppettier"
	"fkirill.org/eink-meteo-station/renderable"
	"fkirill.org/eink-meteo-station/renderable/utils"
	"github.com/rotisserie/eris"
	"image"
	"math/rand"
	"strconv"
	"text/template"
	"time"
)

type PressureRenderable interface {
	renderable.Renderable
}

func NewHAPressureView(
	rect image.Rectangle,
	timeProvider utils.TimeProvider,
	cfg config.ConfigApi,
	envData environment.EnvironmentDataProvider,
) PressureRenderable {
	var pressureWidgetSize = image.Point{X: rect.Dx(), Y: rect.Dy()}
	raster := make([]byte, pressureWidgetSize.X*pressureWidgetSize.Y, pressureWidgetSize.X*pressureWidgetSize.Y)
	for i := range raster {
		raster[i] = 0xff
	}
	tmpl, err := template.New("pressure").Parse(pressureTemplate)
	if err != nil {
		panic(eris.ToString(eris.Wrap(err, "Error parsing pressure template"), true))
	}
	return &pressureView{
		config:                 cfg,
		envData:                envData,
		size:                   pressureWidgetSize,
		offset:                 rect.Min,
		nextRedrawTime:         timeProvider.UtcNow(),
		raster:                 raster,
		pressure:               &environment.PressureData{},
		timeProvider:           timeProvider,
		pressureParsedTemplate: tmpl,
	}
}

type pressureView struct {
	config                 config.ConfigApi
	envData                environment.EnvironmentDataProvider
	size                   image.Point
	offset                 image.Point
	nextRedrawTime         time.Time
	raster                 []byte
	pressure               *environment.PressureData
	timeProvider           utils.TimeProvider
	pressureParsedTemplate *template.Template
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
	pressure, err := p.envData.GetPressure()
	pressureNeedsRedraw := false
	if err != nil {
		if !p.pressure.Warning {
			p.pressure.Warning = true
			pressureNeedsRedraw = true
		}
	} else {
		if *pressure != *p.pressure {
			p.pressure = pressure
			pressureNeedsRedraw = true
		}
	}
	if pressureNeedsRedraw {
		html, err := p.generatePressureHtml(p.pressure)
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

func (p *pressureView) DisplayMode() uint8 {
	return clib.A2_Mode
}

func (p *pressureView) generatePressureHtml(pressureData *environment.PressureData) (string, error) {
	buffer := bytes.Buffer{}
	err := p.pressureParsedTemplate.Execute(&buffer, pressureData)
	if err != nil {
		return "", err
	}
	return string(buffer.Bytes()), nil
}
