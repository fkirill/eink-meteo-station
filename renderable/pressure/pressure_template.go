package pressure

var pressureTemplate = `<html lang="en">
<head>
    <link rel="stylesheet" href="fonts.css"/>
</head>
<body style="margin: 0">
	<div style="padding: 67px; display: inline">
		<div>
			<span style="border-radius: 40px; border: 4px solid; font-size: 80px; padding: 13px; font-family: verily; font-weight: bold">Pressure</span>
			{{if .Warning}}<img src="{{ .WarningPng }}" width="67" height="67"/>{{end}}
		</div>
		<div style="margin-top: 0px">
            <table>
                <tr>
                    <td rowspan=2>
                        <span style="font-size: 133px; font-family: cartograph">{{.PressureInt}}</span>
                        <img src="{{if .PressureRising}}{{ .RisingPng }}{{end}}{{if .PressureFalling}}{{ .FallingPng }}{{end}}{{if .PressureSteady}}{{ .SteadyPng }}{{end}}" width="30" height="30"/>
                        <!-- spacer -->
                        <span style="marginLeft: 30px">&nbsp;</span>
                    </td>
                    <td style="font-size: 60px; font-family: cartograph">mm</td>
                </tr>
                <tr>
                    <td style="font-size: 60px; font-family: cartograph">Hg</td>
                </tr>
            </table>
		</div>
		<div style="margin-top: 0px">
			<span style="font-size: 60px; font-family: cartograph">{{.PressureDeltaInt}}</span>
			<span style="font-size: 60px; font-family: cartograph">.{{.PressureDeltaFrac}}</span>
			<span style="font-size: 40px; font-family: cartograph">&nbsp;{{if .PressureAboveNorm}}above{{end}}{{if .PressureBelowNorm}}below{{end}} norm</span>
		</div>
	</div>
</body>
</html>`
