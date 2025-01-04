package temperature

var temperatureTemplate = `<html lang="en">
<head>
    <link rel="stylesheet" href="fonts.css"/>
</head>
<body style="margin: 0">
	<div style="padding: 67px; display: inline">
		<div>
			<span style="border-radius: 40px; border: 4px solid; font-size: 80px; padding: 13px; font-family: 'verily'; font-weight: bold">{{.Title}}</span>
			{{if .Warning}}<img src="warning.png" width="67" height="67"/>{{end}}
		</div>
		<div style="margin-top: 27px">
			<img src="thermometer.png" width="67" height="67"/>
			<span style="font-size: 133px; font-family: 'cartograph'">{{.TemperatureInt}}</span>
			<span style="font-size: 80px; font-family: 'cartograph'">.{{.TemperatureFrac}}</span>
			<img src="{{if .TemperatureRising}}rising.png{{end}}{{if .TemperatureFalling}}falling.png{{end}}{{if .TemperatureSteady}}steady.png{{end}}" width="30" height="30"/>
		</div>
		<div style="margin-top: 0px">
			<img src="humidity.png" width="67" height="67"/>
            {{if .HundredPercentHumidity}}<span style="font-size: 133px; font-family: cartograph">100</span>{{else}}<span style="font-size: 133px; font-family: cartograph">{{.HumidityInt}}</span>
			<span style="font-size: 80px; font-family: cartograph">.{{.HumidityFrac}}</span>{{end}}
			<img src="{{if .HumidityRising}}rising.png{{end}}{{if .HumidityFalling}}falling.png{{end}}{{if .HumiditySteady}}steady.png{{end}}" width="30" height="30"/>
		</div>
	</div>
</body>
</html>`
