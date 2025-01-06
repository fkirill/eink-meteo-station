package sunset_sunrise

var sunsetSunriseTemplate = `<html lang="en">
<head>
    <link rel="stylesheet" href="fonts.css"/>
</head>
<body style="margin: 0">
	<div style="padding: 67px; display: inline">
		<div>
			<span style="border-radius: 40px; border: 4px solid; font-size: 80px; padding: 13px; font-family: verily; font-weight: bold">Daylight</span>
		</div>
		<div style="margin-top: 27px">
			<img src="{{ .SunrisePng}}" width="67" height="67"/>
			<span style="font-size: 100px; font-family: cartograph">{{.SunriseTime}}</span>
		</div>
		<div style="margin-top: 0px">
			<img src="{{ .SunsetPng }}" width="67" height="67"/>
			<span style="font-size: 100px; font-family: cartograph">{{.SunsetTime}}</span>
		</div>
	</div>
</body>
</html>`
