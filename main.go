package main

import (
	"fkirill.org/eink-meteo-station/di"
	"fkirill.org/eink-meteo-station/systemd"
	"github.com/jessevdk/go-flags"
	"github.com/rotisserie/eris"
	"os"
)

func main() {
	_, err := flags.Parse(&Options{})

	if err != nil {
		if !flags.WroteHelp(err) {
			panic(err)
		}
		panic("Error parsing command line. See previous messages.")
	}
}

type Options struct {
	Service *ServiceOptions `command:"service" description:"Register, start, stop eink-meteo-station as a service"`
	Run     *RunOptions     `command:"run" description:"Run eink-meteo-station in the foreground, requires sudo"`
}

type ServiceOptions struct {
	Install     *ServiceInstallOptions     `command:"install" description:"installs a SystemD service, requires a sudo"`
	Start       *StartServiceOptions       `command:"start" description:"starts a SystemD service, requires a sudo"`
	Stop        *StopServiceOptions        `command:"stop" description:"starts a SystemD service, requires a sudo"`
	IsRunning   *IsServiceRunningOptions   `command:"is-running" description:"checks if the SystemD service is running"`
	IsInstalled *IsServiceInstalledOptions `command:"is-installed" description:"checks if the SystemD service is installed"`
}

type ServiceInstallOptions struct {
	Vcom                       float64 `short:"v" long:"vcom" description:"e-ink screen driving voltage, can be found on a e-ink screen wiring" required:"true"`
	NoWebServer                bool    `short:"n" long:"no-web-server" description:"don't start web server"`
	WebServerListenOnInterface string  `short:"i" long:"interface" description:"interface web server will listen on (empty for all interfaces)" default:""`
	WebServerListenOnPort      uint16  `short:"p" long:"port" description:"port web server will listen on" default:"8080"`
}

func checkRootOrFail(svc systemd.SystemServiceInstaller) error {
	isRoot := svc.CheckRoot()
	if !isRoot {
		return eris.New("Not running under root, this command requires sudo")
	}
	return nil
}

func (s *ServiceInstallOptions) Execute(args []string) error {
	svc := di.GetServiceInstaller()
	if err := checkRootOrFail(svc); err != nil {
		return err
	}
	err := svc.InstallService(s.Vcom, s.NoWebServer, s.WebServerListenOnInterface, s.WebServerListenOnPort)
	if err != nil {
		return eris.Wrap(err, "Error installing the service")
	}
	return nil
}

type StartServiceOptions struct{}

func (s *StartServiceOptions) Execute(args []string) error {
	svc := di.GetServiceInstaller()
	if err := checkRootOrFail(svc); err != nil {
		return err
	}
	err := svc.StartService()
	if err != nil {
		return eris.Wrap(err, "Error starting the service")
	}
	return nil
}

type StopServiceOptions struct{}

func (s *StopServiceOptions) Execute(args []string) error {
	svc := di.GetServiceInstaller()
	if err := checkRootOrFail(svc); err != nil {
		return err
	}
	err := svc.StopService()
	if err != nil {
		return eris.Wrap(err, "Error stopping the service")
	}
	return nil
}

type IsServiceRunningOptions struct{}

func (s *IsServiceRunningOptions) Execute(args []string) error {
	svc := di.GetServiceInstaller()
	isRunning, err := svc.IsServiceRunning()
	if err != nil {
		return eris.Wrap(err, "Error checking if the service is running")
	}
	if isRunning {
		println("Service is running")
	} else {
		println("Service is not running")
	}
	return nil
}

type IsServiceInstalledOptions struct{}

func (s *IsServiceInstalledOptions) Execute(args []string) error {
	svc := di.GetServiceInstaller()
	isRunning, err := svc.IsServiceInstalled()
	if err != nil {
		return eris.Wrap(err, "Error checking if the service is installed")
	}
	if isRunning {
		println("Service is installed")
	} else {
		println("Service is not installed")
	}
	return nil
}

type RunOptions struct {
	Vcom                       float64 `short:"v" long:"vcom" description:"e-ink screen driving voltage, can be found on a e-ink screen wiring" required:"true"`
	NoWebServer                bool    `short:"n" long:"no-web-server" description:"don't start web server"`
	WebServerListenOnInterface string  `short:"i" long:"interface" description:"interface web server will listen on (empty for all interfaces)" default:""`
	WebServerListenOnPort      uint16  `short:"p" long:"port" description:"port web server will listen on" default:"8080"`
}

func (s *RunOptions) Execute(args []string) error {
	if s.NoWebServer {
		svc, err := di.GetMeteoStation(s.Vcom)
		if err != nil {
			return eris.Wrap(err, "Error initializing meteo station objects")
		}
		err = svc.Run()
		if err != nil {
			return eris.Wrap(err, "Error running meteo-station main loop")
		}
	} else {
		meteoAndWeb, err := di.GetMeteoStationAndWebServer(s.Vcom)
		if err != nil {
			return eris.Wrap(err, "Error initializing meteo station objects")
		}
		go func() {
			webErr := meteoAndWeb.WebServer.Start()
			if webErr != nil {
				println(eris.ToString(eris.Wrap(webErr, "Error running web server"), true))
				os.Exit(1)
			}
		}()
		err = meteoAndWeb.MainLoop.Run()
		if err != nil {
			return eris.Wrap(err, "Error running meteo-station main loop")
		}
	}
	return nil
}
