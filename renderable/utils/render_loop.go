package utils

import (
	"fkirill.org/eink-meteo-station/clib"
	"fkirill.org/eink-meteo-station/config"
	"fkirill.org/eink-meteo-station/eink"
	"image"
	"time"
)

type RenderLoop interface {
	Run() error
}

type renderLoop struct {
	first           bool
	timeProvider    TimeProvider
	einkScreen      eink.EInkScreen
	multiRenderable MultiRenderable
	configApi       config.ConfigApi
	diffRenderer    DiffRenderer
}

func (r *renderLoop) Run() error {
	w, h := r.einkScreen.GetScreenDimensions()
	screenSize := image.Point{X: int(w), Y: int(h)}

	currentDate := r.timeProvider.LocalNow().Truncate(24 * time.Hour)
	// main loop
	for {
		timeToNextDraw := r.multiRenderable.NextRedrawDateTimeUtc().Sub(r.timeProvider.UtcNow())
		if timeToNextDraw.Nanoseconds() > 0 {
			time.Sleep(timeToNextDraw)
		}
		if r.configApi.GetRedrawAll() {
			clib.EPD_IT8951_Clear_Refresh(uint16(screenSize.X), uint16(screenSize.Y), r.einkScreen.GetBufferAddress(), clib.INIT_Mode)
			r.configApi.ResetRedrawAll()
			r.multiRenderable.RedrawNow()
		}
		err := r.multiRenderable.Render()
		if err != nil {
			panic(err)
		}
		displayMode := r.multiRenderable.DisplayMode()
		r.multiRenderable.RedrawFinished()
		rect, err := r.diffRenderer.SingleRenderPass(r.multiRenderable.Raster())
		if err != nil {
			println("Diff render failed")
			panic(err)
		}
		if rect.Empty() {
			continue
		}
		// full redraw at midnight
		date := r.timeProvider.LocalNow().Truncate(24 * time.Hour)
		if date != currentDate {
			currentDate = date
			displayMode = clib.GC16_Mode
			rect = image.Rectangle{Max: screenSize}
		}
		if r.first {
			displayMode = clib.GC16_Mode
			rect = image.Rectangle{Max: screenSize}
			r.first = false
		}
		if r.configApi.GetSimpleRefresh() {
			clib.EPD_IT8951_Clear_Refresh(uint16(screenSize.X), uint16(screenSize.Y), r.einkScreen.GetBufferAddress(), clib.INIT_Mode)
			displayMode = clib.GC16_Mode
			rect = image.Rectangle{Max: screenSize}
			r.configApi.ResetSimpleRefresh()
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
			rectBuffer = r.multiRenderable.Raster()
		} else {
			rectBuffer = CutRectangle(rect, screenSize, r.multiRenderable.Raster())
		}

		compressed, err := CompressRasterTo4bpp(
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
		err = r.einkScreen.WriteScreenAreaRefreshMode(displayRect, compressed, displayMode)
		if err != nil {
			panic(err)
		}
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

func NewRenderLoop(
	timeProvider TimeProvider,
	einkScreen eink.EInkScreen,
	multiRenderable MultiRenderable,
	cfg config.ConfigApi,
	diffRenderer DiffRenderer,
) RenderLoop {
	return &renderLoop{
		first:           true,
		timeProvider:    timeProvider,
		einkScreen:      einkScreen,
		multiRenderable: multiRenderable,
		configApi:       cfg,
		diffRenderer:    diffRenderer,
	}
}
