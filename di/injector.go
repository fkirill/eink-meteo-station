//go:build wireinject
// +build wireinject

package di

import (
	"fkirill.org/eink-meteo-station/renderable/utils"
	"fkirill.org/eink-meteo-station/systemd"
	"fkirill.org/eink-meteo-station/webui"
	"github.com/google/wire"
)

type Injector struct {
	MainLoop  utils.RenderLoop
	WebServer webui.WebServer
}

func GetMeteoStationAndWebServer(vcom float64) (*Injector, error) {
	wire.Build(
		configModule,
		dataModule,
		einkModule,
		renderableModule,
		webModule,
		utilModule,
		wire.Struct(new(Injector), "*"),
	)
	return nil, nil
}

func GetMeteoStation(vcom float64) (utils.RenderLoop, error) {
	wire.Build(
		configModule,
		dataModule,
		einkModule,
		renderableModule,
		utilModule,
	)
	return nil, nil
}

func GetServiceInstaller() systemd.SystemServiceInstaller {
	wire.Build(serviceModule)
	return nil
}
