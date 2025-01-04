package temperature

import (
	"bytes"
	"fkirill.org/eink-meteo-station/clib"
	"fkirill.org/eink-meteo-station/data/environment"
	"fkirill.org/eink-meteo-station/puppettier"
	"fkirill.org/eink-meteo-station/renderable"
	"fkirill.org/eink-meteo-station/renderable/utils"
	"github.com/rotisserie/eris"
	"html/template"
	"image"
	"math/rand"
	"strconv"
	"time"
)

type TemperatureHumidityRenderable interface {
	renderable.Renderable
}

func NewHATemperatureView(
	rect image.Rectangle,
	timeProvider utils.TimeProvider,
	envProvider environment.EnvironmentDataProvider,
) (TemperatureHumidityRenderable, error) {
	rasterSize := rect.Dx() * rect.Dy()
	raster := make([]byte, rasterSize, rasterSize)
	for i := range raster {
		raster[i] = 0xff
	}
	tmpl, err := template.New("temperature").Parse(temperatureTemplate)
	if err != nil {
		return nil, eris.Wrap(err, "Cannot parse template")
	}
	return &temperatureView{
		temperatureParsedTemplate: tmpl,
		envProvider:               envProvider,
		size:                      rect.Size(),
		offset:                    rect.Min,
		nextRedrawTime:            timeProvider.UtcNow(),
		raster:                    raster,
		timeProvider:              timeProvider,
		cachedInside:              &environment.TemperatureHumidityData{},
		cachedOutside:             &environment.TemperatureHumidityData{},
	}, nil
}

type temperatureView struct {
	temperatureParsedTemplate *template.Template
	envProvider               environment.EnvironmentDataProvider
	size                      image.Point
	offset                    image.Point
	nextRedrawTime            time.Time
	raster                    []byte
	timeProvider              utils.TimeProvider
	cachedInside              *environment.TemperatureHumidityData
	cachedOutside             *environment.TemperatureHumidityData
}

func (t *temperatureView) RedrawNow() {
	t.nextRedrawTime = t.timeProvider.UtcNow()
}

func (t *temperatureView) generateTemperatureHtml(temperatureData *environment.TemperatureHumidityData) (string, error) {
	buffer := bytes.Buffer{}
	err := t.temperatureParsedTemplate.Execute(&buffer, temperatureData)
	if err != nil {
		return "", err
	}
	return string(buffer.Bytes()), nil
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
	signleTempViewSize := image.Point{t.size.X / 2, t.size.Y}
	inside, err := t.envProvider.GetInsideTemperatureHumidity()
	if inside == nil {
		inside = &environment.TemperatureHumidityData{}
	}
	insideNeedsRedraw := false
	if err != nil {
		if !t.cachedInside.Warning {
			t.cachedInside.Warning = true
			insideNeedsRedraw = true
		}
	} else {
		if *inside != *t.cachedInside {
			t.cachedInside = inside
			insideNeedsRedraw = true
		}
	}
	outside, err := t.envProvider.GetOutsideTemperatureHumidity()
	if outside == nil {
		outside = &environment.TemperatureHumidityData{}
	}
	outsideNeedsRedraw := false
	if err != nil {
		if !t.cachedOutside.Warning {
			t.cachedOutside.Warning = true
			outsideNeedsRedraw = true
		}
	} else {
		if *outside != *t.cachedOutside {
			t.cachedOutside = outside
			outsideNeedsRedraw = true
		}
	}
	if insideNeedsRedraw {
		html, err := t.generateTemperatureHtml(t.cachedInside)
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
		html, err := t.generateTemperatureHtml(t.cachedOutside)
		if err != nil {
			return err
		}
		img, err := puppettier.RenderInPuppeteer(html, "temperature_"+strconv.FormatInt(time.Now().Unix()+100, 10), signleTempViewSize)
		if err != nil {
			return err
		}
		utils.DrawImage(t.raster, t.size, image.Point{X: signleTempViewSize.X, Y: 0}, img, signleTempViewSize)
	}
	return nil
}

func (t *temperatureView) DisplayMode() uint8 {
	return clib.A2_Mode
}
