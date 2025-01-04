package systemd

import (
	"bytes"
	"errors"
	"github.com/rotisserie/eris"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"text/template"
	"time"
)

const serviceDir = "/usr/lib/systemd/system"
const serviceFile = "eink-meteo-station.service"

type SystemServiceInstaller interface {
	IsServiceInstalled() (bool, error)
	IsServiceRunning() (bool, error)
	InstallService(vcom float64, noWebServer bool, listenInterface string, listenPort uint16) error
	StartService() error
	StopService() error
	CheckRoot() bool
}

type systemServiceInstaller struct{}

func (s systemServiceInstaller) CheckRoot() bool {
	user, exists := os.LookupEnv("USER")
	if !exists {
		return false
	}
	if user != "root" {
		return false
	}
	return true
}

// systemctl exit status
// status 0 - service running or active
// status 3 - service disabled or inactive
// status 4 - service not found

func (s systemServiceInstaller) IsServiceInstalled() (bool, error) {
	status, _, err := getStatusCode("sudo", "systemctl", "status", serviceFile)
	if err != nil {
		return false, eris.Wrap(err, "Error getting service status")
	}
	if status != 0 && status != 3 {
		return false, nil
	}
	return true, nil
}

func (s systemServiceInstaller) IsServiceRunning() (bool, error) {
	status, _, err := getStatusCode("sudo", "systemctl", "status", serviceFile)
	if err != nil {
		return false, eris.Wrap(err, "Error getting service status")
	}
	return status == 0, nil
}

func getStatusCode(commandLine ...string) (int, string, error) {
	procChan := make(chan bool, 1)
	cmd := exec.Command(commandLine[0], commandLine[1:]...)
	if cmd.Err != nil {
		return 0, "", eris.Wrap(cmd.Err, "Error starting sudo systemctl")
	}
	var outputBuf []byte
	var runErr error
	go func() {
		outputBuf, runErr = cmd.CombinedOutput()
		procChan <- true
	}()
	timeoutTicker := time.NewTicker(10 * time.Second)
	defer timeoutTicker.Stop()
	select {
	case <-timeoutTicker.C:
		err := cmd.Process.Kill()
		if err != nil {
			err = eris.Wrap(err, "Error killing the sudo systemctl process")
		}
		timeoutErr := eris.New("Timed out waiting for sudo systemctl process")
		if err != nil {
			timeoutErr = errors.Join(timeoutErr, err)
		}
		return 0, "", timeoutErr
	case <-procChan:
		timeoutTicker.Stop()
	}
	if runErr != nil && !strings.HasPrefix(runErr.Error(), "exit status") {
		return 0, "", eris.Wrapf(runErr, "Error starting command %v", commandLine)
	}
	if !cmd.ProcessState.Exited() {
		return 0, "", eris.New("process sudo systemctl didn't exit as planned")
	}
	return cmd.ProcessState.ExitCode(), string(outputBuf), nil
}

type TemplateData struct {
	EinkServiceExecutablePath string
	EinkServiceDirPath        string
	Parameters                string
}

func (s systemServiceInstaller) InstallService(vcom float64, noWebServer bool, listenInterface string, listenPort uint16) error {
	parameters := []string{"-v", strconv.FormatFloat(vcom, 'g', 4, 64)}
	if noWebServer {
		parameters = append(parameters, "-n")
	}
	if listenInterface != "" {
		parameters = append(parameters, "-i", listenInterface)
	}
	if listenPort != 8080 {
		parameters = append(parameters, "-p", strconv.Itoa(int(listenPort)))
	}
	tmpl, err := template.New("service").Parse(`
[Unit]
Description=E-Ink Meteo-station service
After=network.target auditd.service

[Service]
ExecStart={{.EinkServiceExecutablePath}} run {{.Parameters}}
KillMode=process
Restart=on-failure
Type=notify
RootDirectory={{.EinkServiceDirPath}}

[Install]
WantedBy=multi-user.target
`)
	if err != nil {
		return eris.Wrap(err, "Error parsing template")
	}
	curDir, err := os.Getwd()
	if err != nil {
		return eris.Wrap(err, "Error getting current directory")
	}
	executable := path.Join(curDir, path.Base(os.Args[0]))
	tmplData := &TemplateData{
		EinkServiceExecutablePath: path.Base(executable),
		EinkServiceDirPath:        path.Dir(executable),
		Parameters:                strings.Join(parameters, " "),
	}
	buf := &bytes.Buffer{}
	err = tmpl.Execute(buf, tmplData)
	if err != nil {
		return eris.Wrap(err, "Error executing template")
	}
	err = os.WriteFile(path.Join(serviceDir, serviceFile), buf.Bytes(), 0644)
	if err != nil {
		return eris.Wrap(err, "Error writing service content. Did you run under sudo?")
	}
	status, output, err := getStatusCode("sudo", "systemctl", "reload", serviceFile)
	if err != nil {
		return eris.Wrapf(err, "Error reloading service %s", serviceFile)
	}
	status, output, err = getStatusCode("sudo", "systemctl", "enable", serviceFile)
	if err != nil {
		return eris.Wrapf(err, "Error enabling service %s", serviceFile)
	}
	if status != 0 {
		return eris.Errorf("Error enabling service %s\nOutput:\n%s", serviceFile, output)
	}
	status, output, err = getStatusCode("sudo", "systemctl", "status", serviceFile)
	if err != nil {
		return eris.Wrapf(err, "Error getting service status %s", serviceFile)
	}
	if status != 0 && status != 3 {
		return eris.Errorf("Error getting service status %s, exist code:%d\nOutput:\n%s", serviceFile, status, output)
	}
	return nil
}

func (s systemServiceInstaller) StartService() error {
	status, output, err := getStatusCode("sudo", "systemctl", "status", serviceFile)
	if err != nil {
		return eris.Wrapf(err, "Error getting service status %s", serviceFile)
	}
	if status == 4 {
		return eris.Errorf("Service %s is not registered", serviceFile)
	}
	if status != 0 {
		return eris.Errorf("Unexpected return status from sudo systemctl for service %s\nOutput:\n%s", serviceFile, output)
	}
	status, output, err = getStatusCode("sudo", "systemctl", "restart", serviceFile)
	if err != nil {
		return eris.Wrapf(err, "Error restarting service %s", serviceFile)
	}
	if status != 0 {
		return eris.Errorf("Unexpected return status from sudo systemctl for service %s\nOutput:\n%s", serviceFile, output)
	}
	return nil
}

func (s systemServiceInstaller) StopService() error {
	status, output, err := getStatusCode("sudo", "systemctl", "status", serviceFile)
	if err != nil {
		return eris.Wrapf(err, "Error getting service status %s", serviceFile)
	}
	if status == 4 {
		return eris.Errorf("Service %s is not registered", serviceFile)
	}
	if status != 0 {
		return eris.Errorf("Unexpected return status from sudo systemctl for service %s\nOutput:\n%s", serviceFile, output)
	}
	status, output, err = getStatusCode("sudo", "systemctl", "stop", serviceFile)
	if err != nil {
		return eris.Wrapf(err, "Error stopping service %s", serviceFile)
	}
	if status != 0 {
		return eris.Errorf("Unexpected return status from sudo systemctl for service %s\nOutput:\n%s", serviceFile, output)
	}
	return nil
}

func NewSystemServiceInstaller() SystemServiceInstaller {
	return &systemServiceInstaller{}
}
