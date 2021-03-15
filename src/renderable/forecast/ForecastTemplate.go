package forecast

import (
	"bytes"
	"html/template"
	"strconv"
)

type dailyForecast struct {
	DayOfMonth string // two characters
	DayOfWeek string // three characters
	Month string // three characters
	MinTemp string // two characters
	MaxTemp string // two characters
	AmountOfRain string // two characters
	MaxWind string // two characters
	WeatherType int
}

type forecastTable struct {
	Days []*dailyForecast
}

var forecastTemplate = `<html lang="en">
<head>
    <link rel="stylesheet" href="fonts.css"/>
</head>
<body style="margin: 0">
  <table>
    <thead>
      <tr>
        <td><span style="border-radius: 40px; border: 4px solid; font-size: 40px; padding: 13px; font-family: 'verily'; font-weight: bold">Forecast</span></td>
        {{range .Days}}<td><div style="text-align: center; font-size: 80px; font-family: 'cartograph'; margin-left:20px; margin-right: 20px">{{.DayOfMonth}}</div><div style="text-align: center; font-size: 40px; font-family: 'cartograph'">{{.DayOfWeek}}</div></td>{{end}}
      </tr>
    </thead>
    <tbody>
      <tr>
        <td style="font-size: 50px; font-family: 'cartograph'">t&nbsp;max</td>
        {{range .Days}}<td><div style="text-align: center; font-size: 80px; font-family: 'cartograph'">{{.MaxTemp}}</div></td>{{end}}
      </tr>
      <tr>
        <td style="font-size: 50px; font-family: 'cartograph'">t&nbsp;min</td>
        {{range .Days}}<td><div style="text-align: center; font-size: 80px; font-family: 'cartograph'">{{.MinTemp}}</div></td>{{end}}
      </tr>
      <tr>
        <td style="font-size: 50px; font-family: 'cartograph'">rain</td>
        {{range .Days}}<td><div style="text-align: center; font-size: 80px; font-family: 'cartograph'">{{.AmountOfRain}}</div></td>{{end}}
      </tr>
    <tbody>
  </table>
</body>
</html>`

var forecastParsedTemplate *template.Template

func init() {
	tmpl, err := template.New("forecast").Parse(forecastTemplate)
	if err != nil {
		panic("Cannot parse template: " + err.Error())
	}
	forecastParsedTemplate = tmpl
}

func GenerateForecastHtml(forecastData *ForecastData) (string, error) {
	forecastTable := convertToTemplateFormat(forecastData.Days)
	buffer := bytes.Buffer{}
	err := forecastParsedTemplate.Execute(&buffer, forecastTable)
	if err != nil {
		return "", err
	}
	return string(buffer.Bytes()), nil
}

func convertToTemplateFormat(days []ForecastDataDay) *forecastTable {
	res := make([]*dailyForecast, 0)
	for _, day := range days {
		daily := &dailyForecast{
			DayOfMonth:   strconv.Itoa(day.Date.Day()),
			DayOfWeek:    day.Date.Weekday().String()[:3],
			Month:        day.Date.Month().String()[:3],
			MinTemp:      strconv.Itoa(int(day.MinTemp)),
			MaxTemp:      strconv.Itoa(int(day.MaxTemp)),
			AmountOfRain: strconv.Itoa(int(day.ExpectedRainAmountMm)),
			MaxWind:      strconv.Itoa(int(day.MaxWindKmh)),
			WeatherType:  0,
		}
		res = append(res, daily)
	}
	return &forecastTable{
		Days: res,
	}
}
