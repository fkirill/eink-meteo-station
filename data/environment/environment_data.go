package environment

import (
	"errors"
	"fkirill.org/eink-meteo-station/config"
	"fkirill.org/eink-meteo-station/data/ha"
	"fkirill.org/eink-meteo-station/images"
	"math"
	"strconv"
	"time"
)

type PressureData struct {
	Warning           bool   // display an error sign
	PressureInt       string // three digits
	PressureFrac      string // one digit
	PressureRising    bool   // One of the three must be true
	PressureSteady    bool   // the other two must be false
	PressureFalling   bool
	PressureAboveNorm bool   // one of the two must be true, the other must be false
	PressureBelowNorm bool   // when delta == 0.0, it is considered "above" for display purposes
	PressureDeltaInt  string // two characters, first may be space if delta <10 mmHg
	PressureDeltaFrac string // one digit
	WarningPng        string
	RisingPng         string
	FallingPng        string
	SteadyPng         string
}

type TemperatureHumidityData struct {
	Title                  string
	Warning                bool   // display an error sign
	TemperatureInt         string // two digits
	TemperatureFrac        string // one digit
	TemperatureRising      bool   // One of the three must be true
	TemperatureFalling     bool   // the other two must be false
	TemperatureSteady      bool
	HumidityInt            string // two digits
	HumidityFrac           string // one digit
	HumidityRising         bool   // One of the three must be true
	HumidityFalling        bool   // the other two must be false
	HumiditySteady         bool
	HundredPercentHumidity bool
	WarningPng             string
	ThermometerPng         string
	RisingPng              string
	FallingPng             string
	SteadyPng              string
	HumidityPng            string
}

type EnvironmentDataProvider interface {
	GetInsideTemperatureHumidity() (*TemperatureHumidityData, error)
	GetOutsideTemperatureHumidity() (*TemperatureHumidityData, error)
	GetPressure() (*PressureData, error)
}

type environmentDataProvider struct {
	config config.ConfigApi
	haApi  ha.HomeAssistantApi
}

func (e *environmentDataProvider) GetInsideTemperatureHumidity() (*TemperatureHumidityData, error) {
	insideTemperatureSensorName := e.config.GetInternalTemperatureSensorName()
	insideHumiditySensorName := e.config.GetInternalHumiditySensorName()
	return e.getTemperatureHumidity("Inside", insideTemperatureSensorName, insideHumiditySensorName)
}

func (e *environmentDataProvider) GetOutsideTemperatureHumidity() (*TemperatureHumidityData, error) {
	outsideTemperatureSensorName := e.config.GetExternalTemperatureSensorName()
	outsideHumiditySensorName := e.config.GetExternalHumiditySensorName()
	return e.getTemperatureHumidity("Outside", outsideTemperatureSensorName, outsideHumiditySensorName)
}

func (e *environmentDataProvider) GetPressure() (*PressureData, error) {
	pressureSensorName := e.config.GetPressureSensorName()
	pressure, err := e.haApi.DownloadSensorValueFromHA(pressureSensorName)
	if err != nil {
		return nil, err
	}
	pressureVal, err := strconv.ParseFloat(pressure, 64)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	pressureVal /= hPaToMmHgCoeff
	pressureHistory, err := e.haApi.DownloadSensorHistoryFromHA(pressureSensorName, now.Add(-60*time.Minute), now, false)
	if err != nil {
		return nil, err
	}
	pressureNumericSeries, err := convertToNumericSeries(pressureHistory)
	if err != nil {
		return nil, err
	}
	pressureSlope := slope(pressureNumericSeries) * 3600 / hPaToMmHgCoeff
	pressureRising := pressureSlope >= 1.0
	pressureFalling := pressureSlope <= -1.0
	pressureSteady := !(pressureRising || pressureFalling)
	pressureDelta := pressureVal - normalPressureMmHg
	pressureAboveNorm := pressureDelta >= 0.0
	pressureInt := strconv.Itoa(int(math.Trunc(pressureVal)))                  // always yields three digits
	pressureDeltaInt := strconv.Itoa(int(math.Trunc(math.Abs(pressureDelta)))) // one or two digits
	if len(pressureDeltaInt) == 1 {
		pressureDeltaInt = " " + pressureDeltaInt
	}
	return &PressureData{
		Warning:           false,
		PressureInt:       pressureInt,
		PressureFrac:      formatFrac(pressureVal),
		PressureRising:    pressureRising,
		PressureSteady:    pressureSteady,
		PressureFalling:   pressureFalling,
		PressureAboveNorm: pressureAboveNorm,
		PressureBelowNorm: !pressureAboveNorm,
		PressureDeltaInt:  pressureDeltaInt,
		PressureDeltaFrac: formatFrac(math.Abs(pressureDelta)),
		WarningPng:        images.Warning_png_src,
		RisingPng:         images.Rising_png_src,
		FallingPng:        images.Falling_png_src,
		SteadyPng:         images.Steady_png_src,
	}, nil
}

func NewEnvironmentDataProvider(config config.ConfigApi, haApi ha.HomeAssistantApi) EnvironmentDataProvider {
	return &environmentDataProvider{config, haApi}
}

const hPaToMmHgCoeff = 1.33
const normalPressureMmHg = 760.0

func (e *environmentDataProvider) getTemperatureHumidity(title, temperatureSensorName, humiditySensorName string) (*TemperatureHumidityData, error) {
	temp, err := e.haApi.DownloadSensorValueFromHA(temperatureSensorName)
	if err != nil {
		return nil, err
	}
	tempVal, err := strconv.ParseFloat(temp, 64)
	if err != nil {
		return nil, err
	}
	humidity, err := e.haApi.DownloadSensorValueFromHA(humiditySensorName)
	if err != nil {
		return nil, err
	}
	humidityVal, err := strconv.ParseFloat(humidity, 64)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	tempHistory, err := e.haApi.DownloadSensorHistoryFromHA(temperatureSensorName, now.Add(-30*time.Minute), now, false)
	if err != nil {
		return nil, err
	}
	humidityHistory, err := e.haApi.DownloadSensorHistoryFromHA(humiditySensorName, now.Add(-30*time.Minute), now, false)
	if err != nil {
		return nil, err
	}
	tempNumericSeries, err := convertToNumericSeries(tempHistory)
	if err != nil {
		return nil, err
	}
	humidityNumericSeries, err := convertToNumericSeries(humidityHistory)
	if err != nil {
		return nil, err
	}
	tempSlope := slope(tempNumericSeries) * 3600
	humiditySlope := slope(humidityNumericSeries) * 3600
	tempRising := tempSlope >= 0.3
	tempFalling := tempSlope <= -0.3
	tempSteady := !(tempRising || tempFalling)
	humidityRising := humiditySlope >= 0.3
	humidityFalling := humiditySlope <= -0.3
	humiditySteady := !(humidityRising || humidityFalling)
	hundredPercentHumidity := humidityVal > 99.9
	return &TemperatureHumidityData{
		Title:                  title,
		Warning:                false,
		TemperatureInt:         formatInt(tempVal),
		TemperatureFrac:        formatFrac(tempVal),
		TemperatureRising:      tempRising,
		TemperatureFalling:     tempFalling,
		TemperatureSteady:      tempSteady,
		HumidityInt:            formatInt(humidityVal),
		HumidityFrac:           formatFrac(humidityVal),
		HumidityRising:         humidityRising,
		HumidityFalling:        humidityFalling,
		HumiditySteady:         humiditySteady,
		HundredPercentHumidity: hundredPercentHumidity,
		WarningPng:             images.Warning_png_src,
		ThermometerPng:         images.Thermometer_png_src,
		RisingPng:              images.Rising_png_src,
		FallingPng:             images.Falling_png_src,
		SteadyPng:              images.Steady_png_src,
		HumidityPng:            images.Humidity_png_src,
	}, nil
}

func formatInt(val float64) string {
	truncated := int(math.Trunc(val))
	if truncated <= -10 || truncated >= 100 {
		return ".."
	}
	if truncated < 0 {
		return "-" + strconv.Itoa(-truncated)
	}
	if truncated < 10 {
		return " " + strconv.Itoa(truncated)
	}
	return strconv.Itoa(truncated)
}

func formatFrac(val float64) string {
	firstFracDigit := int(math.Trunc(math.Trunc(val*10.0) - math.Trunc(val)*10.0))
	if firstFracDigit < 0 {
		firstFracDigit = -firstFracDigit
	}
	return strconv.Itoa(firstFracDigit)
}

// Calculates a non-vertical slope coefficient for a given set of points.
// It uses liner regression formula described here:
// https://en.wikipedia.org/wiki/Simple_linear_regression#Fitting_the_regression_line
// the result is the approximate  change of the value per second calculated over the data provided
// you may want to multiply by 3600 to get the hourly change
func slope(res []*ha.NumericHistoryValue) float64 {
	sum_t := int64(0)
	sum_c := 0.0
	for _, e := range res {
		sum_t += e.Timestamp.Unix()
		sum_c += e.Value
	}
	avg_t := sum_t / int64(len(res))
	avg_c := sum_c / float64(len(res))
	num := float64(0)
	denomt := float64(0)
	for _, e := range res {
		t := e.Timestamp.Unix()
		num += float64(t-avg_t) * (e.Value - avg_c)
		denomt += float64((t - avg_t) * (t - avg_t))
	}
	beta := num / denomt
	return beta
}

func convertToNumericSeries(source []*ha.HomeAssistantHistoryItem) ([]*ha.NumericHistoryValue, error) {
	if source == nil {
		return nil, errors.New("source was nil")
	}
	res := make([]*ha.NumericHistoryValue, 0)

	for i := range source {
		if source[i].Value == "unavailable" || source[i].Value == "unknown" {
			continue
		}
		val, err := strconv.ParseFloat(source[i].Value, 64)
		if err != nil {
			return nil, err
		}
		res = append(res, &ha.NumericHistoryValue{
			Timestamp: source[i].Timestamp,
			Value:     val,
		})
	}
	return res, nil
}
