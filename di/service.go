package di

import (
	"fkirill.org/eink-meteo-station/systemd"
	"github.com/google/wire"
)

func provideServiceInstaller() systemd.SystemServiceInstaller {
	return systemd.NewSystemServiceInstaller()
}

var serviceModule = wire.NewSet(
	provideServiceInstaller,
)
