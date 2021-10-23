package webui

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"renderable/config"
	"strconv"
	"time"
)

type renderData struct {
	Message                   string
	InternalTemperatureSensor string
	ExternalTemperatureSensor string
	InternalHumiditySensor    string
	ExternalHumiditySensor    string
	PressureSensor            string
	SpecialDays               []config.SpecialDayOrInterval
}

var configPageTemplateText = `
<html>
<head>
  <title>RPI Meteostation configuration</title>
</head>
<body>
  {{.Message}}
  <h1>Commands</h1>
  <div>
    <form action="/" method="post">
      <input type="hidden" name="command" value="redraw_all"/>
      <button type="submit">Redraw all</button>
    </form>
  </div>
  <h1>HomeAssistant sensor names</h1>
  <div>
    <form action="/" method="post">
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
      <input type="hidden" name="command" value="set_sensor_names"/>
      <button type="submit">Update sensors</button>
    </form>
  </div>
  <h1>Special days</h1>
  <div>
    <form action="/" method="post">
{{ range .SpecialDays }}
      Index: {{.Index}}
      <br/>
      Id: <input type="text" name="special_days.{{.Index}}.id" value="{{.Id}}" />
      <br />
      Display text: <input type="text" name="special_days.{{.Index}}.display_text" value="{{.DisplayText}}" />
      <br />
      Type:
      <select name="special_days.{{.Index}}.type" value="{{.Type}}">
        <option value="once_off">Once off</option>
        <option value="annual">Annual</option>
        <option value="interval">Interval</option>
      </select>
      <br />
      <input type="checkbox" id="public_holiday" name="public_holiday" value="true"{{ if .IsPublicHoliday }} checked{{end}} />
      <label for="public_holiday">Public holiday</label>
      <br />
      <input type="checkbox" id="school_holiday" name="school_holiday" value="true"{{ if .IsSchoolHoliday }} checked{{end}} />
      <label for="public_holiday">School holiday</label>
      <br />
      Start date day: <input type="number" min="1" max="31" name="special_days.{{.Index}}.start_day" value="{{.StartDateDay}}" />
      Month: 
      <select name="special_days.{{.Index}}.start_month" value="{{.StartDateMonth}}">
        <option value="1"{{ if eq .StartDateMonth 1 }} selected{{end}}>January</option>
        <option value="2"{{ if eq .StartDateMonth 2 }} selected{{end}}>February</option>
        <option value="3"{{ if eq .StartDateMonth 3 }} selected{{end}}>March</option>
        <option value="4"{{ if eq .StartDateMonth 4 }} selected{{end}}>April</option>
        <option value="5"{{ if eq .StartDateMonth 5 }} selected{{end}}>May</option>
        <option value="6"{{ if eq .StartDateMonth 6 }} selected{{end}}>June</option>
        <option value="7"{{ if eq .StartDateMonth 7 }} selected{{end}}>July</option>
        <option value="8"{{ if eq .StartDateMonth 8 }} selected{{end}}>August</option>
        <option value="9"{{ if eq .StartDateMonth 9 }} selected{{end}}>September</option>
        <option value="10"{{ if eq .StartDateMonth 10 }} selected{{end}}>October</option>
        <option value="11"{{ if eq .StartDateMonth 11 }} selected{{end}}>November</option>
        <option value="12"{{ if eq .StartDateMonth 12 }} selected{{end}}>December</option>
      </select>
      Year: <input type="number" min="2021" max="2100" name="special_days.{{.Index}}.start_year" value="{{.StartDateYear}}" />
      <br />
      End date day: <input type="number" min="1" max="31" name="special_days.{{.Index}}.end_day" value="{{.EndDateDay}}" />
      Month: 
      <select name="special_days.{{.Index}}.end_month" value="{{.EndDateMonth}}">
        <option value="1"{{ if eq .EndDateMonth 1 }} selected{{end}}>January</option>
        <option value="2"{{ if eq .EndDateMonth 2 }} selected{{end}}>February</option>
        <option value="3"{{ if eq .EndDateMonth 3 }} selected{{end}}>March</option>
        <option value="4"{{ if eq .EndDateMonth 4 }} selected{{end}}>April</option>
        <option value="5"{{ if eq .EndDateMonth 5 }} selected{{end}}>May</option>
        <option value="6"{{ if eq .EndDateMonth 6 }} selected{{end}}>June</option>
        <option value="7"{{ if eq .EndDateMonth 7 }} selected{{end}}>July</option>
        <option value="8"{{ if eq .EndDateMonth 8 }} selected{{end}}>August</option>
        <option value="9"{{ if eq .EndDateMonth 9 }} selected{{end}}>September</option>
        <option value="10"{{ if eq .EndDateMonth 10 }} selected{{end}}>October</option>
        <option value="11"{{ if eq .EndDateMonth 11 }} selected{{end}}>November</option>
        <option value="12"{{ if eq .EndDateMonth 12 }} selected{{end}}>December</option>
      </select>
      Year: <input type="number" min="2021" max="2100" name="special_days.{{.Index}}.end_year" value="{{.EndDateYear}}" />
      <br/>
      -------------------------------------
      <br/>
{{ end }}
      <button type="submit">Update special days</button>
      <input type="hidden" name="command" value="set_special_days"/>
    </form>
    <form action="/" method="post">
      <button type="submit">Add special day</button>
      <input type="hidden" name="command" value="add_special_day"/>
    </form>
    <form action="/" method="post">
      <select name="special_day_index">
{{ range .SpecialDays }}
        <option value="{{.Index}}">{{.Id}}</option>
{{ end }}
      </select>
      <button type="submit">Remove special day</button>
      <input type="hidden" name="command" value="remove_special_day"/>
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
	SetSpecialDays(specialDays []config.SpecialDayOrInterval)
	GetSpecialDays() []config.SpecialDayOrInterval
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
var specialDays []config.SpecialDayOrInterval

func (ws *webServer) mainHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		command := r.FormValue("command")
		if len(command) > 0 {
			if command == "redraw_all" {
				ws.redrawAll()
			} else if command == "set_sensor_names" {
				ws.setSensorNames(r)
			} else if command == "set_special_days" {
				ws.setSpecialDays(r)
			} else if command == "add_special_day" {
				ws.addSpecialDay()
			} else if command == "remove_special_day" {
				ws.removeSpecialDay(r)
			} else {
				message = fmt.Sprintf("Commande not recognized: %s", command)
			}
		} else {
			message = "Error: command not provided"
		}
	}
	data := &renderData{
		Message:                   message,
		InternalTemperatureSensor: ws.configApi.GetInternalTemperatureSensorName(),
		ExternalTemperatureSensor: ws.configApi.GetExternalTemperatureSensorName(),
		InternalHumiditySensor:    ws.configApi.GetInternalHumiditySensorName(),
		ExternalHumiditySensor:    ws.configApi.GetExternalHumiditySensorName(),
		PressureSensor:            ws.configApi.GetPressureSensorName(),
		SpecialDays:               specialDays,
	}
	err := tmpl.Execute(w, data)
	if err != nil {
		log.Printf("Error writing page output %v", err)
	}
	message = ""
}

func (ws *webServer) setSpecialDays(r *http.Request) {
	message = ""
	for i := range specialDays {
		idStr := r.FormValue(fmt.Sprintf("special_days.%d.id", specialDays[i].Index))
		if len(idStr) == 0 {
			message += "; Warning: special day id must not be empty"
		}
		displayTextStr := r.FormValue(fmt.Sprintf("special_days.%d.display_text", specialDays[i].Index))
		if len(displayTextStr) == 0 {
			message += "; Warning: special day display text must not be empty"
		}
		typeStr := r.FormValue(fmt.Sprintf("special_days.%d.type", specialDays[i].Index))
		if typeStr != "once_off" && typeStr != "annual" && typeStr != "interval" {
			message += fmt.Sprintf("; Error: unrecognized special day type: '%s'", typeStr)
			continue
		}
		isInterval := typeStr == "interval"
		isAnnual := typeStr == "annual"

		specialDays[i].Id = idStr
		specialDays[i].DisplayText = displayTextStr
		specialDays[i].Type = typeStr
		startDayStr := r.FormValue(fmt.Sprintf("special_days.%d.start_day", specialDays[i].Index))
		startDay, err := strconv.Atoi(startDayStr)
		if err != nil {
			message += fmt.Sprintf("; Error: cannot parse start date day '%s'", startDayStr)
			continue
		}
		if startDay < 1 || startDay > 31 {
			message += fmt.Sprintf("; Warning: start date day %d is outside [1,31] range", startDay)
		}
		specialDays[i].StartDateDay = startDay
		startMonthStr := r.FormValue(fmt.Sprintf("special_days.%d.start_month", specialDays[i].Index))
		startMonth, err := strconv.Atoi(startMonthStr)
		if err != nil {
			message += fmt.Sprintf("; Error: cannot parse start date month '%s'", startMonthStr)
			continue
		}
		if startMonth < 1 || startMonth > 12 {
			message += fmt.Sprintf("; Warning: start date month %d is outside [1,12] range", startMonth)
		}
		specialDays[i].StartDateMonth = startMonth
		if !isAnnual {
			startYearStr := r.FormValue(fmt.Sprintf("special_days.%d.start_year", specialDays[i].Index))
			startYear, err := strconv.Atoi(startYearStr)
			if err != nil {
				message += fmt.Sprintf("; Error: cannot parse start date year '%s'", startYearStr)
				continue
			}
			if startYear < 2020 || startYear > 2050 {
				message += fmt.Sprintf("; Warning: start date year %d is outside [2020,2050] range", startYear)
			}
			specialDays[i].StartDateYear = startYear
		}
		if isInterval {
			endDayStr := r.FormValue(fmt.Sprintf("special_days.%d.end_day", specialDays[i].Index))
			endDay, err := strconv.Atoi(endDayStr)
			if err != nil {
				message += fmt.Sprintf("; Error: cannot parse end date day '%s'", endDayStr)
				continue
			}
			specialDays[i].EndDateDay = endDay
			endMonthStr := r.FormValue(fmt.Sprintf("special_days.%d.end_month", specialDays[i].Index))
			endMonth, err := strconv.Atoi(endMonthStr)
			if err != nil {
				message += fmt.Sprintf("; Error: cannot parse end date month '%s'", endMonthStr)
				continue
			}
			specialDays[i].EndDateMonth = endMonth
			endYearStr := r.FormValue(fmt.Sprintf("special_days.%d.end_year", specialDays[i].Index))
			specialDays[i].EndDateYear, err = strconv.Atoi(endYearStr)
			if err != nil {
				message += fmt.Sprintf("; Error: cannot parse end date year '%s'", endYearStr)
				continue
			}
			startDttm := time.Date(
				specialDays[i].StartDateYear,
				time.Month(specialDays[i].StartDateMonth),
				specialDays[i].StartDateDay,
				0, 0, 0, 0, time.UTC,
			)
			endDttm := time.Date(
				specialDays[i].EndDateYear,
				time.Month(specialDays[i].EndDateMonth),
				specialDays[i].EndDateDay,
				0, 0, 0, 0, time.UTC,
			)
			if startDttm.After(endDttm) {
				message += "; Warning: start date must be before end date"
			}
		}
		isPublicHolidayStr := r.FormValue("public_holiday")
		isSchoolHolidayStr := r.FormValue("school_holiday")
		specialDays[i].IsPublicHoliday = isPublicHolidayStr == "true"
		specialDays[i].IsSchoolHoliday = isSchoolHolidayStr == "true"
	}
	ws.configApi.SetSpecialDays(specialDays)
}

func (ws *webServer) redrawAll() {
	ws.configApi.RedrawAll()
	message = "Full redraw initiated"
}

func (ws *webServer) removeSpecialDay(r *http.Request) {
	specialDayIndexStr := r.FormValue("special_day_index")
	specialDayIndex, err := strconv.Atoi(specialDayIndexStr)
	if err == nil {
		newSpecialDays := make([]config.SpecialDayOrInterval, 0)
		for _, sd := range specialDays {
			if sd.Index != specialDayIndex {
				newSpecialDays = append(newSpecialDays, sd)
			}
		}
		specialDays = newSpecialDays
		updateSpecialDayIndices()
		message = "Special day added"
	} else {
		message = fmt.Sprintf("error: cannot convert '%s' to int", specialDayIndexStr)
	}
}

func (ws *webServer) addSpecialDay() {
	today := time.Now().In(time.UTC)
	specialDays = append(specialDays, config.SpecialDayOrInterval{
		StartDateDay: today.Day(),
		StartDateMonth: int(today.Month()),
		StartDateYear: today.Year(),
		EndDateDay: today.Day(),
		EndDateMonth: int(today.Month()),
		EndDateYear: today.Year(),
	})
	updateSpecialDayIndices()
	message = "Special day added"
}

func (ws *webServer) setSensorNames(r *http.Request) {
	ws.configApi.SetInternalTemperatureSensorName(r.FormValue("internal_temperature_sensor"))
	ws.configApi.SetExternalTemperatureSensorName(r.FormValue("external_temperature_sensor"))
	ws.configApi.SetInternalHumiditySensorName(r.FormValue("internal_humidity_sensor"))
	ws.configApi.SetExternalHumiditySensorName(r.FormValue("external_humidity_sensor"))
	ws.configApi.SetPressureSensorName(r.FormValue("pressure_sensor"))
	ws.configApi.RedrawAll()
	message = "Sensors updated successfully, full redraw initiated"
}

func updateSpecialDayIndices() {
	for i := range specialDays {
		specialDays[i].Index = i + 1
	}
}

func NewWebServer(configApi ConfigApi) WebServer {
	return &webServer{configApi: configApi}
}

func (ws *webServer) Start() {
	http.HandleFunc("/", ws.mainHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
