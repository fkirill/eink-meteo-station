package di

import (
	"fkirill.org/eink-meteo-station/config"
	"fkirill.org/eink-meteo-station/eink"
	"fkirill.org/eink-meteo-station/renderable/utils"
	"github.com/google/wire"
	"image"
)

func provideRenderLoop(
	timeProvider utils.TimeProvider,
	einkScreen eink.EInkScreen,
	multiRenderable utils.MultiRenderable,
	cfg config.ConfigApi,
	diffRenderer utils.DiffRenderer,
) utils.RenderLoop {
	return utils.NewRenderLoop(timeProvider, einkScreen, multiRenderable, cfg, diffRenderer)
}

func provideTimeProvider() utils.TimeProvider {
	return utils.NewTimeProvider()
}

func provideDiffRenderer(screen eink.EInkScreen) utils.DiffRenderer {
	w, h := screen.GetScreenDimensions()
	return utils.NewDiffRenderer(image.Point{X: int(w), Y: int(h)})
}

var utilModule = wire.NewSet(
	provideRenderLoop,
	provideTimeProvider,
	provideDiffRenderer,
)
