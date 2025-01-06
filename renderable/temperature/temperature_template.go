package temperature

var temperatureTemplate = `<html lang="en">
<head>
    <link rel="stylesheet" href="fonts.css"/>
</head>
<body style="margin: 0">
	<div style="padding: 67px; display: inline">
		<div>
			<span style="border-radius: 40px; border: 4px solid; font-size: 80px; padding: 13px; font-family: verily; font-weight: bold">{{.Title}}</span>
			{{if .Warning}}<img src="{{ .WarningPng }}" width="67" height="67"/>{{end}}
		</div>
		<div style="margin-top: 27px">
			<img src="{{ .ThermometerPng }}" width="67" height="67"/>
			<span style="font-size: 133px; font-family: cartograph">{{.TemperatureInt}}</span>
			<span style="font-size: 80px; font-family: cartograph">.{{.TemperatureFrac}}</span>
			<img src="{{if .TemperatureRising}}{{ .RisingPng }}{{end}}{{if .TemperatureFalling}}{ .FallingPng }}{{end}}{{if .TemperatureSteady}}{{ .SteadyPng }}{{end}}" width="30" height="30"/>
		</div>
		<div style="margin-top: 0px">
			<img src="{{ .HumidityPng }}" width="67" height="67"/>
            {{if .HundredPercentHumidity}}<span style="font-size: 133px; font-family: cartograph">100</span>{{else}}<span style="font-size: 133px; font-family: cartograph">{{.HumidityInt}}</span>
			<span style="font-size: 80px; font-family: cartograph">.{{.HumidityFrac}}</span>{{end}}
			<img src="{{if .HumidityRising}}{{ .RisingPng }}{{end}}{{if .HumidityFalling}}{{ .FallingPng }}{{end}}{{if .HumiditySteady}}{{ .SteadyPng }}{{end}}" width="30" height="30"/>
		</div>
	</div>
</body>
</html>`
