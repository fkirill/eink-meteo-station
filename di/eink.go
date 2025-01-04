package di

import (
	"fkirill.org/eink-meteo-station/eink"
	"github.com/google/wire"
)

func provideEinkScreenProvider() eink.EinkScreenProvider {
	return eink.NewEinkScreenProvider()
}

func provideEinkScreen(vcom float64, provider eink.EinkScreenProvider) (eink.EInkScreen, error) {
	return provider.NewEinkScreen(vcom)
}

var einkModule = wire.NewSet(
	provideEinkScreenProvider,
	provideEinkScreen,
)
