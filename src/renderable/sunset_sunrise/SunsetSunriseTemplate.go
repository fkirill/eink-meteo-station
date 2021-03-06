package sunset_sunrise

import (
	"bytes"
	"html/template"
)

var sunsetSunriseTemplate = `<html lang="en">
<head>
    <link rel="stylesheet" href="fonts.css"/>
</head>
<body style="margin: 0">
	<div style="padding: 67px; display: inline">
		<div>
			<span style="border-radius: 40px; border: 4px solid; font-size: 80px; padding: 13px; font-family: 'verily'; font-weight: bold">Daylight</span>
		</div>
		<div style="margin-top: 27px">
			<img src="sunrise.png" width="67" height="67"/>
			<span style="font-size: 100px; font-family: 'cartograph'">{{.SunriseTime}}</span>
		</div>
		<div style="margin-top: 0px">
			<img src="sunset.png" width="67" height="67"/>
			<span style="font-size: 100px; font-family: cartograph">{{.SunsetTime}}</span>
		</div>
	</div>
</body>
</html>`

var sunsetSunriseParsedTemplate *template.Template

func init() {
	tmpl, err := template.New("sunsetSunrise").Parse(sunsetSunriseTemplate)
	if err != nil {
		panic("Cannot parse template: " + err.Error())
	}
	sunsetSunriseParsedTemplate = tmpl
}

func GenerateSunriseHtml(sunsetSunriseData *SunsetSunriseData) (string, error) {
	buffer := bytes.Buffer{}
	err := sunsetSunriseParsedTemplate.Execute(&buffer, sunsetSunriseData)
	if err != nil {
		return "", err
	}
	return string(buffer.Bytes()), nil
}