package webui

import (
	"html/template"
	"log"
	"net/http"
)

type renderData struct {
	Message string
	InternalTemperatureSensor string
	ExternalTemperatureSensor string
	InternalHumiditySensor string
	ExternalHumiditySensor string
	PressureSensor string
}

var configPageTemplateText = `
<html>
<head>
  <title>RPI Meteostation configuration</title>
</head>
<body>
  {{.Message}}
  <div>
    <form action="/redraw" method="get">
      <button type="submit">Redraw all</button>
    </form>
  </div>
  <div>
    <form action="/update_sensors" method="post">
      <div>
        Internal temperature sensor name: <input type="text" name="internal_temperature_sensor" value="{{.InternalTemperatureSensor}}"/>
      </div>
      <div>
        External temperature sensor name: <input type="text" name="external_temperature_sensor" value="{{.ExternalTemperatureSensor}}"/>
      </div>
      <div>
        Internal humidity sensor name: <input type="text" name="internal_humidity_sensor" value="{{.InternalHumiditySensor}}"/>
      </div>
      <div>
        External humidity sensor name: <input type="text" name="external_humidity_sensor" value="{{.ExternalHumiditySensor}}"/>
      </div>
      <div>
        Pressure sensor name: <input type="text" name="pressure_sensor" value="{{.PressureSensor}}"/>
      </div>
      <button type="submit">Update sensors</button>
    </form>
  </div>
</body>
</html>
`

type ConfigApi interface {
	RedrawAll()
	SetInternalTemperatureSensorName(sensorName string)
	SetInternalHumiditySensorName(sensorName string)
	SetExternalTemperatureSensorName(sensorName string)
	SetExternalHumiditySensorName(sensorName string)
	SetPressureSensorName(sensorName string)
	GetInternalTemperatureSensorName() string
	GetInternalHumiditySensorName() string
	GetExternalTemperatureSensorName() string
	GetExternalHumiditySensorName() string
	GetPressureSensorName() string
}

var tmpl *template.Template

func init() {
	var err error
	tmpl, err = template.New("config").Parse(configPageTemplateText)
	if err != nil {
		log.Fatalf("cannot parse html template: %v", err)
	}
}

type WebServer interface {
	Start()
}

type webServer struct {
	configApi ConfigApi
}

var message string

func (ws* webServer) mainHandler(w http.ResponseWriter, _ *http.Request) {
	data := &renderData{
		Message:                   message,
		InternalTemperatureSensor: ws.configApi.GetInternalTemperatureSensorName(),
		ExternalTemperatureSensor: ws.configApi.GetExternalTemperatureSensorName(),
		InternalHumiditySensor:    ws.configApi.GetInternalHumiditySensorName(),
		ExternalHumiditySensor:    ws.configApi.GetExternalHumiditySensorName(),
		PressureSensor:            ws.configApi.GetPressureSensorName(),
	}
	err := tmpl.Execute(w, data)
	if err != nil {
		log.Printf("Error writing page output %v", err)
	}
	message = ""
}

func (ws* webServer) redrawHandler(w http.ResponseWriter, r *http.Request) {
	message = "Full redraw initiated"
	ws.configApi.RedrawAll()
	ws.mainHandler(w, r)
}

func (ws* webServer) updateSensorsHandler(w http.ResponseWriter, r *http.Request) {
	ws.configApi.SetInternalTemperatureSensorName(r.FormValue("internal_temperature_sensor"))
	ws.configApi.SetExternalTemperatureSensorName(r.FormValue("external_temperature_sensor"))
	ws.configApi.SetInternalHumiditySensorName(r.FormValue("internal_humidity_sensor"))
	ws.configApi.SetExternalHumiditySensorName(r.FormValue("external_humidity_sensor"))
	ws.configApi.SetPressureSensorName(r.FormValue("pressure_sensor"))
	message = "Sensors updated successfully, full redraw initiated"
	ws.configApi.RedrawAll()
	ws.mainHandler(w, r)
}

func NewWebServer(configApi ConfigApi) WebServer {
	return &webServer{configApi: configApi}
}

func (ws* webServer) Start() {
	http.HandleFunc("/redraw", ws.redrawHandler)
	http.HandleFunc("/update_sensors", ws.updateSensorsHandler)
	http.HandleFunc("/", ws.mainHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
