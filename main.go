package main

import (
	"bytes"
	"fkirill.org/eink-meteo-station/clib"
	"fkirill.org/eink-meteo-station/renderable"
	"fkirill.org/eink-meteo-station/renderable/calendar"
	"fkirill.org/eink-meteo-station/renderable/clock"
	"fkirill.org/eink-meteo-station/renderable/config"
	"fkirill.org/eink-meteo-station/renderable/forecast"
	"fkirill.org/eink-meteo-station/renderable/pressure"
	"fkirill.org/eink-meteo-station/renderable/sunset_sunrise"
	"fkirill.org/eink-meteo-station/renderable/temperature"
	"fkirill.org/eink-meteo-station/renderable/utils"
	"fkirill.org/eink-meteo-station/webui"
	"image"
	"image/png"
	"os"
	"time"
)

var pathToDisplayDriverProcess = "~/eink-screen-driver/IT8951"

var first = false

func main() {
	//timeProvider := utils.NewTestTimeProvider(time.Now().Truncate(24 * time.Hour).Add(-10 * time.Second))
	timeProvider := utils.NewTimeProvider()

	//Init the BCM2835 Device
	screen, err := NewEInkScreen(-1.2)
	if err != nil {
		panic(err)
	}
	w, h := screen.GetScreenDimensions()
	screenSize := image.Point{X: int(w), Y: int(h)}
	clockView, err := clock.NewClockRenderable(image.Point{}, timeProvider)
	if err != nil {
		panic(err)
	}
	calendarView := calendar.NewCalendarRenderable(image.Point{Y: 280}, image.Point{X: 962, Y: 1120}, timeProvider)
	temperatureView := temperature.NewHATemperatureView(
		image.Point{1000, 0},
		timeProvider,
		config.GetInternalTemperatureSensor,
		config.GetExternalTemperatureSensor,
		config.GetInternalHumiditySensor,
		config.GetExternalHumiditySensor,
	)
	pressureView := pressure.NewHAPressureView(image.Point{1000, 500}, timeProvider, config.GetPressureSensor)
	latitude, longitude := config.GetDaylightCoordinates()
	daylightView := sunset_sunrise.NewSunriseSunsetRenderable(image.Point{1450, 500}, timeProvider, latitude, longitude)
	forecastView := forecast.NewForecastRenderable(image.Point{X: 1000, Y: 900}, timeProvider)
	multiRenderable, err := renderable.NewMultiRenderable(
		image.Point{},
		screenSize,
		[]renderable.Renderable{forecastView, calendarView, clockView, temperatureView, pressureView, daylightView},
		//[]renderable.Renderable{forecastView, calendarView, temperatureView, pressureView, daylightView},
		false)

	configApi := newConfigApi(multiRenderable, calendarView)
	ws := webui.NewWebServer(configApi)
	go ws.Start()

	currentDate := timeProvider.LocalNow().Truncate(24 * time.Hour)
	diffRenderer := renderable.NewDiffRenderer(screenSize)
	// main loop
	for {
		timeToNextDraw := multiRenderable.NextRedrawDateTimeUtc().Sub(timeProvider.UtcNow())
		if timeToNextDraw.Nanoseconds() > 0 {
			time.Sleep(timeToNextDraw)
		}
		err = multiRenderable.Render()
		if err != nil {
			panic(err)
		}
		displayMode := multiRenderable.DisplayMode()
		multiRenderable.RedrawFinished()
		rect, err := diffRenderer.SingleRenderPass(multiRenderable.Raster())
		if err != nil {
			println("Diff render failed")
			panic(err)
		}
		if rect.Empty() {
			continue
		}
		// full redraw at midnight
		date := timeProvider.LocalNow().Truncate(24 * time.Hour)
		if date != currentDate {
			currentDate = date
			displayMode = clib.GC16_Mode
			rect = image.Rectangle{Max: screenSize}
		}
		if first {
			displayMode = clib.GC16_Mode
			rect = image.Rectangle{Max: screenSize}
			first = false
		}
		if config.GetSimpleRefresh() {
			clib.EPD_IT8951_Clear_Refresh(uint16(screenSize.X), uint16(screenSize.Y), screen.GetBufferAddress(), clib.INIT_Mode)
			displayMode = clib.GC16_Mode
			rect = image.Rectangle{Max: screenSize}
			config.ResetSimpleRefresh()
		}
		// important: expand the range slightly to make sure that each row occupies even number of bytes
		// given that we're talking 4bpp compact encoding it means that the rectangle should start and end
		// at the X coordinates multiple of 4.
		if rect.Min.X%4 != 0 {
			rect.Min.X -= rect.Min.X % 4
		}
		if rect.Max.X%4 != 0 {
			rect.Max.X += 4 - rect.Max.X%4
		}
		fullScreen := rect.Size() == screenSize
		var rectBuffer []byte
		if fullScreen {
			rectBuffer = multiRenderable.Raster()
		} else {
			rectBuffer = renderable.CutRectangle(rect, screenSize, multiRenderable.Raster())
		}

		compressed, err := renderable.CompressRasterTo4bpp(
			image.Rectangle{Max: image.Point{X: rect.Dx(), Y: rect.Dy()}},
			image.Point{X: rect.Dx(), Y: rect.Dy()},
			rectBuffer,
			true,
		)
		if err != nil {
			println("Image compression failed")
			panic(err)
		}
		displayRect := alignRectangles(rect, screenSize.X)
		err = screen.WriteScreenAreaRefreshMode(displayRect, compressed, displayMode)
		if err != nil {
			panic(err)
		}
	}
}

func writePng(compressed []byte, size image.Point) {
	img := &image.Gray{
		Pix:    compressed,
		Stride: size.X,
		Rect:   image.Rectangle{Max: image.Point{X: size.X - 1, Y: size.Y - 1}},
	}
	buf := &bytes.Buffer{}
	err := png.Encode(buf, img)
	if err != nil {
		panic(err)
	}
	err = os.WriteFile(time.Now().Format(time.RFC3339)+".png", buf.Bytes(), 0644)
	if err != nil {
		panic(err)
	}
}

func alignRectangles(r image.Rectangle, width int) image.Rectangle {
	r.Min.X = width - r.Min.X
	r.Max.X = width - r.Max.X
	temp := r.Min.X
	r.Min.X = r.Max.X
	r.Max.X = temp
	return r
}

type configApi struct {
	multiRenderable    renderable.Renderable
	calendarRenderable renderable.Renderable
}

func (c configApi) SetSpecialDays(specialDays []*config.SpecialDayOrInterval) {
	config.SetSpecialDays(specialDays)
	c.calendarRenderable.RedrawNow()
}

func (c configApi) GetSpecialDays() []*config.SpecialDayOrInterval {
	return config.GetSpecialDays()
}

func (c configApi) GetInternalTemperatureSensorName() string {
	return config.GetInternalTemperatureSensor()
}

func (c configApi) GetInternalHumiditySensorName() string {
	return config.GetInternalHumiditySensor()
}

func (c configApi) GetExternalTemperatureSensorName() string {
	return config.GetExternalTemperatureSensor()
}

func (c configApi) GetExternalHumiditySensorName() string {
	return config.GetExternalHumiditySensor()
}

func (c configApi) GetPressureSensorName() string {
	return config.GetPressureSensor()
}

func (c configApi) RedrawAll() {
	c.multiRenderable.RedrawNow()
	first = true
}

func (c configApi) SetInternalTemperatureSensorName(sensorName string) {
	config.SetInternalTemperatureSensor(sensorName)
}

func (c configApi) SetInternalHumiditySensorName(sensorName string) {
	config.SetInternalHumiditySensor(sensorName)
}

func (c configApi) SetExternalTemperatureSensorName(sensorName string) {
	config.SetExternalTemperatureSensor(sensorName)
}

func (c configApi) SetExternalHumiditySensorName(sensorName string) {
	config.SetExternalHumiditySensor(sensorName)
}

func (c configApi) SetPressureSensorName(sensorName string) {
	config.SetPressureSensor(sensorName)
}

func (c configApi) GetDaylightCoordinates() (float64, float64) {
	return config.GetDaylightCoordinates()
}

func newConfigApi(multiRenderable, calendarRenderable renderable.Renderable) webui.ConfigApi {
	return &configApi{multiRenderable: multiRenderable, calendarRenderable: calendarRenderable}
}

type EInkScreen interface {
	GetScreenDimensions() (uint16, uint16)
	GetBufferAddress() uint32
	WriteScreenAreaRefreshMode(area image.Rectangle, raster []byte, mode uint8) error
}

type einkScreen struct {
	buffer         []uint8
	panelW, panelH uint16
	bufferAddress  uint32
	fwVersion      string
	lutVersion     string
}

func (e *einkScreen) WriteScreenAreaRefreshMode(area image.Rectangle, raster []byte, mode uint8) error {
	clib.EPD_IT8951_4bp_Refresh_Mode(
		raster,
		uint16(area.Min.X),
		uint16(area.Min.Y),
		uint16(area.Dx()),
		uint16(area.Dy()),
		false,
		e.bufferAddress,
		false,
		mode,
	)
	return nil
}

func (e *einkScreen) GetScreenDimensions() (uint16, uint16) {
	return e.panelW, e.panelH
}

func (e *einkScreen) GetBufferAddress() uint32 {
	return e.bufferAddress
}

func NewEInkScreen(VCom float64) (EInkScreen, error) {
	if clib.DEV_Module_Init() != 0 {
		panic("Failed to initialize eink screen")
	}
	VCOM := uint16(1200)
	Dev_Info := clib.EPD_IT8951_Init(VCOM)
	clib.Epd_Mode(1)
	clib.EPD_IT8951_Clear_Refresh(Dev_Info.Panel_W, Dev_Info.Panel_H, Dev_Info.Memory_Addr, clib.INIT_Mode)
	bufSize := Dev_Info.Panel_W * Dev_Info.Panel_H / 2
	return &einkScreen{
		buffer:        make([]uint8, bufSize, bufSize),
		panelW:        Dev_Info.Panel_W,
		panelH:        Dev_Info.Panel_H,
		bufferAddress: Dev_Info.Memory_Addr,
		fwVersion:     wordsToString(Dev_Info.FW_Version[:]),
		lutVersion:    wordsToString(Dev_Info.LUT_Version[:]),
	}, nil
}

func wordsToString(words []uint16) string {
	buffer := bytes.Buffer{}
	for _, w := range words {
		buffer.WriteByte(byte(w << 8))
		buffer.WriteByte(byte(w & 0xff))
	}
	buf := buffer.Bytes()
	var l int
	for i := range buf {
		if buf[i] == 0 {
			l = i
			break
		}
	}
	if l == 0 {
		return ""
	}
	return string(buf[0:l])
}
