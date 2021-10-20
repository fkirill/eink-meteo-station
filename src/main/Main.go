package main

import (
	"bytes"
	"image"
	"io"
	"os/exec"
	"renderable"
	"renderable/calendar"
	"renderable/clock"
	"renderable/forecast"
	"renderable/pressure"
	"renderable/sunset_sunrise"
	"renderable/temperature"
	"renderable/utils"
	"secrets"
	"time"
	"webui"
)

var pathToDisplayDriverProcess = "/home/pi/epaper/bcm2835-1.68/IT8951/IT8951/IT8951"

var first = true

func main() {
	//timeProvider := renderable.NewTestTimeProvider(time.Now().Truncate(24*time.Hour).Add(-10*time.Second))
	timeProvider := utils.NewTimeProvider()

	size := image.Point{X: 1872, Y: 1404}
	calendarView := calendar.NewCalendarRenderable(image.Point{Y: 280}, image.Point{X: 962, Y: 952}, timeProvider)
	clockView, err := clock.NewClockRenderable(image.Point{}, timeProvider)
	if err != nil {
		panic(err)
	}
	temperatureView := temperature.NewHATemperatureView(
		image.Point{1000, 0},
		timeProvider,
		secrets.GetInternalTemperatureSensor,
		secrets.GetExternalTemperatureSensor,
		secrets.GetInternalHumiditySensor,
		secrets.GetExternalHumiditySensor,
	)
	pressureView := pressure.NewHAPressureView(image.Point{1000, 500}, timeProvider, secrets.GetPressureSensor)
	daylightView := sunset_sunrise.NewSunriseSunsetRenderable(image.Point{1450, 500}, timeProvider)
	forecastView := forecast.NewForecastRenderable(image.Point{X: 950, Y: 900}, timeProvider)
	multiRenderable, err := renderable.NewMultiRenderable(
		image.Point{},
		size,
		[]renderable.Renderable{forecastView, calendarView, clockView, temperatureView, pressureView, daylightView},
		false)

	configApi := newConfigApi(multiRenderable)
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
	currentDate := timeProvider.Now().Truncate(24 * time.Hour)
	diffRenderer := renderable.NewDiffRenderer(size)
	// main loop
	for {
		timeToNextDraw := multiRenderable.NextRedrawDateTime().Sub(timeProvider.Now())
		if timeToNextDraw.Nanoseconds() > 0 {
			time.Sleep(timeToNextDraw)
		}
		err = multiRenderable.Render()
		if err != nil {
			panic(err)
		}
		displayMode := multiRenderable.DisplayMode()
		multiRenderable.RedrawFinished()
		rects, err := diffRenderer.SingleRenderPass(multiRenderable.Raster())
		if err != nil {
			println("Diff render failed")
			panic(err)
		}
		if len(rects) == 0 {
			continue
		}
		rects[0] = alignRectangles(rects[0], size.X)
		compressed, err := renderable.CompressRasterTo4bpp(size, multiRenderable.Raster(), true)
		if err != nil {
			println("Image compression failed")
			panic(err)
		}
		if first {
			displayMode = 2
			rects[0] = image.Rectangle{Max: size}
			first = false
		}
		// full redraw at midnight
		date := timeProvider.Now().Truncate(24 * time.Hour)
		if date != currentDate {
			currentDate = date
			displayMode = 2
			rects[0] = image.Rectangle{Max: size}
		}
		buffer := writeImageFrame(size, rects, displayMode, compressed)
		_, err = writer.Write(buffer.Bytes())
		if err != nil {
			panic(err)
		}
	}
}

func writeImageFrame(size image.Point, rects []image.Rectangle, displayMode int, compressed []byte) bytes.Buffer {
	buffer := bytes.Buffer{}
	buffer.WriteByte(0)                            // preamble
	buffer.WriteByte(1)                            // preamble
	buffer.WriteByte(2)                            // preamble
	buffer.WriteByte(3)                            // preamble
	buffer.WriteByte(4)                            // preamble
	buffer.WriteByte(5)                            // preamble
	buffer.WriteByte(6)                            // preamble
	buffer.WriteByte(7)                            // preamble
	buffer.WriteByte(0)                            // shouldExit
	writeShort(&buffer, uint16(size.X))            // image width
	writeShort(&buffer, uint16(size.Y))            // image height
	buffer.WriteByte(byte(len(rects)))             // # of rectangles
	writeShort(&buffer, uint16(rects[0].Min.X))    // rectangle[0].x
	writeShort(&buffer, uint16(rects[0].Min.Y))    // rectangle[0].y
	writeShort(&buffer, uint16(rects[0].Size().X)) // rectangle[0].w
	writeShort(&buffer, uint16(rects[0].Size().Y)) // rectangle[0].h
	buffer.WriteByte(byte(displayMode))            // rectangle[0].displayMode
	buffer.Write(compressed)
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
	multiRenderable renderable.Renderable
}

func (c configApi) GetInternalTemperatureSensorName() string {
	return secrets.GetInternalTemperatureSensor()
}

func (c configApi) GetInternalHumiditySensorName() string {
	return secrets.GetInternalHumiditySensor()
}

func (c configApi) GetExternalTemperatureSensorName() string {
	return secrets.GetExternalTemperatureSensor()
}

func (c configApi) GetExternalHumiditySensorName() string {
	return secrets.GetExternalHumiditySensor()
}

func (c configApi) GetPressureSensorName() string {
	return secrets.GetPressureSensor()
}

func (c configApi) RedrawAll() {
	c.multiRenderable.RedrawNow()
	first = true
}

func (c configApi) SetInternalTemperatureSensorName(sensorName string) {
	secrets.SetInternalTemperatureSensor(sensorName)
}

func (c configApi) SetInternalHumiditySensorName(sensorName string) {
	secrets.SetInternalHumiditySensor(sensorName)
}

func (c configApi) SetExternalTemperatureSensorName(sensorName string) {
	secrets.SetExternalTemperatureSensor(sensorName)
}

func (c configApi) SetExternalHumiditySensorName(sensorName string) {
	secrets.SetExternalHumiditySensor(sensorName)
}

func (c configApi) SetPressureSensorName(sensorName string) {
	secrets.SetPressureSensor(sensorName)
}

func newConfigApi(multiRenderable renderable.Renderable) webui.ConfigApi {
	return &configApi{multiRenderable: multiRenderable}
}
