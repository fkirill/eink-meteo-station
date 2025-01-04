package forecast

type dailyForecast struct {
	DayOfMonth   string // two characters
	DayOfWeek    string // three characters
	Month        string // three characters
	MinTemp      string // two characters
	MaxTemp      string // two characters
	AmountOfRain string // two characters
	AmountOfSnow string // two characters
	MaxWind      string // two characters
	WeatherType  int
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
      <tr>
        <td style="font-size: 50px; font-family: 'cartograph'">snow</td>
        {{range .Days}}<td><div style="text-align: center; font-size: 80px; font-family: 'cartograph'">{{.AmountOfSnow}}</div></td>{{end}}
      </tr>
    <tbody>
  </table>
</body>
</html>`
