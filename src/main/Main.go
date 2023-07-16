package main

import (
	"bytes"
	"image"
	"io"
	"io/ioutil"
	"os/exec"
	"renderable"
	"renderable/calendar"
	"renderable/clock"
	"renderable/config"
	"renderable/forecast"
	"renderable/pressure"
	"renderable/sunset_sunrise"
	"renderable/temperature"
	"renderable/utils"
	"time"
	"webui"
)

var pathToDisplayDriverProcess = "~/eink-screen-driver/IT8951"

var first = true

func main() {
	//timeProvider := renderable.NewTestTimeProvider(time.Now().Truncate(24*time.Hour).Add(-10*time.Second))
	timeProvider := utils.NewTimeProvider()

	size := image.Point{X: 1872, Y: 1404}
	calendarView := calendar.NewCalendarRenderable(image.Point{Y: 280}, image.Point{X: 962, Y: 1120}, timeProvider)
	clockView, err := clock.NewClockRenderable(image.Point{}, timeProvider)
	if err != nil {
		panic(err)
	}
	temperatureView := temperature.NewHATemperatureView(
		image.Point{1000, 0},
		timeProvider,
		config.GetInternalTemperatureSensor,
		config.GetExternalTemperatureSensor,
		config.GetInternalHumiditySensor,
		config.GetExternalHumiditySensor,
	)
	pressureView := pressure.NewHAPressureView(image.Point{1000, 500}, timeProvider, config.GetPressureSensor)
	daylightView := sunset_sunrise.NewSunriseSunsetRenderable(image.Point{1450, 500}, timeProvider)
	forecastView := forecast.NewForecastRenderable(image.Point{X: 950, Y: 900}, timeProvider)
	multiRenderable, err := renderable.NewMultiRenderable(
		image.Point{},
		size,
		[]renderable.Renderable{forecastView, calendarView, clockView, temperatureView, pressureView, daylightView},
		false)

	configApi := newConfigApi(multiRenderable, calendarView)
	ws := webui.NewWebServer(configApi)
	go ws.Start()

	if err != nil {
		panic(err)
	}
	cmd := exec.Command(pathToDisplayDriverProcess)
	writer, err := cmd.StdinPipe()
	if err != nil {
		panic(err)
	}
	err = cmd.Start()
	if err != nil {
		panic(err)
	}
	currentDate := timeProvider.LocalNow().Truncate(24 * time.Hour)
	diffRenderer := renderable.NewDiffRenderer(size)
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
		alignRectangles(rect, size.X)
		// full redraw at midnight
		date := timeProvider.LocalNow().Truncate(24 * time.Hour)
		if date != currentDate {
			currentDate = date
			displayMode = 2
			rect = image.Rectangle{Max: size}
		}
		if first {
			displayMode = 2
			rect = image.Rectangle{Max: size}
			first = false
		}
		compressed, err := renderable.CompressRasterTo4bpp(rect, size, multiRenderable.Raster(), true)
		if err != nil {
			println("Image compression failed")
			panic(err)
		}
		buffer := writeImageFrame(rect, displayMode, compressed)
		ioutil.WriteFile("latest-frame.bin", buffer.Bytes(), 0)
		_, err = writer.Write(buffer.Bytes())
		if err != nil {
			panic(err)
		}
	}
}

func writeImageFrame(rect image.Rectangle, displayMode int, compressed []byte) bytes.Buffer {
	buffer := bytes.Buffer{}
	buffer.WriteByte(0)                     // preamble
	buffer.WriteByte(1)                     // preamble
	buffer.WriteByte(2)                     // preamble
	buffer.WriteByte(3)                     // preamble
	buffer.WriteByte(4)                     // preamble
	buffer.WriteByte(5)                     // preamble
	buffer.WriteByte(6)                     // preamble
	buffer.WriteByte(7)                     // preamble
	buffer.WriteByte(0)                     // shouldExit
	writeShort(&buffer, uint16(rect.Dx()))  // image width
	writeShort(&buffer, uint16(rect.Dy()))  // image height
	writeShort(&buffer, uint16(rect.Min.X)) // image startX
	writeShort(&buffer, uint16(rect.Min.Y)) // image startY
	buffer.WriteByte(byte(displayMode))     // displayMode
	buffer.Write(compressed)                // image data
	return buffer
}

func alignRectangles(r image.Rectangle, width int) image.Rectangle {
	if r.Min.X%2 == 1 {
		r.Min.X--
	}
	if r.Max.X%2 == 1 {
		r.Max.X++
	}
	r.Min.X = width - r.Min.X
	r.Max.X = width - r.Max.X
	temp := r.Min.X
	r.Min.X = r.Max.X
	r.Max.X = temp
	return r
}

func writeShort(w io.Writer, i uint16) {
	_, e := w.Write([]byte{byte(i & 0xff00 >> 8), byte(i & 0x00ff)})
	if e != nil {
		panic("unexpected")
	}
}

type configApi struct {
	multiRenderable    renderable.Renderable
	calendarRenderable renderable.Renderable
}

func (c configApi) SetSpecialDays(specialDays []config.SpecialDayOrInterval) {
	config.SetSpecialDays(specialDays)
	c.calendarRenderable.RedrawNow()
}

func (c configApi) GetSpecialDays() []config.SpecialDayOrInterval {
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

func newConfigApi(multiRenderable, calendarRenderable renderable.Renderable) webui.ConfigApi {
	return &configApi{multiRenderable: multiRenderable, calendarRenderable: calendarRenderable}
}
