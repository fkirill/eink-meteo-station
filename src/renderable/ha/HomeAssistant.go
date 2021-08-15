package ha

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"secrets"
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
}

var authToken = "Bearer " + secrets.GetHAToken()
var homeAssistantProtocol = secrets.GetHAProtocol()
var homeAssistantHost = secrets.GetHAHost()
var homeAssistantPort = secrets.GetHAPort()
var haProtocolHostPort = homeAssistantProtocol + "://" + homeAssistantHost + ":" + strconv.Itoa(int(homeAssistantPort))

func DownloadSensorValueFromHA(sensorId string) (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", haProtocolHostPort+"/api/states/"+sensorId, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Authorization", authToken)
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var objmap map[string]json.RawMessage
	err = json.Unmarshal(body, &objmap)
	if err != nil {
		return "", err
	}
	if _, exists := objmap["state"]; exists {
		byteVal := objmap["state"]
		val := string(byteVal)
		// if the value is quoted, cut off the quotes
		if len(byteVal) >= 2 && byteVal[0] == '"' && byteVal[len(byteVal)-1] == '"' {
			val = string(byteVal[1 : len(byteVal)-1])
		}
		return val, nil
	}
	return "", errors.New("no field 'state' found in json output")
}

type HomeAssistantHistoryItem struct {
	Timestamp time.Time `json:"last_changed"`
	Value     string    `json:"state"`
}

func DownloadSensorHistoryFromHA(sensorId string, startTime, endTime time.Time, significantOnly bool) ([]HomeAssistantHistoryItem, error) {
	client := &http.Client{}
	significantOnlyStr := ""
	if significantOnly {
		significantOnlyStr = "&significant_changes_only"
	}
	startDateTimeStr := startTime.UTC().Format(time.RFC3339)
	endDateTimeStr := endTime.UTC().Format(time.RFC3339)
	req, err := http.NewRequest("GET", "http://192.168.68.110:8123/api/history/period/"+startDateTimeStr+
		"?filter_entity_id="+sensorId+"&end_time"+endDateTimeStr+"&minimal_response"+significantOnlyStr,
		nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", authToken)
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var res [][]HomeAssistantHistoryItem
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}
	// remove quotes
	if len(res) == 0 {
		return nil, fmt.Errorf("no history returned for sensor %s", sensorId)
	}
	for _, e := range res[0] {
		// if the value is quoted, cut off the quotes
		if len(e.Value) >= 2 && e.Value[0] == '"' && e.Value[len(e.Value)-1] == '"' {
			e.Value = e.Value[1 : len(e.Value)-1]
		}
	}
	if err != nil {
		return nil, err
	}
	return res[0], nil
}

type NumericHistoryValue struct {
	Timestamp time.Time
	Value     float64
}

func ConvertToNumericSeries(source []HomeAssistantHistoryItem) ([]NumericHistoryValue, error) {
	if source == nil {
		return nil, errors.New("source was nil")
	}
	res := make([]NumericHistoryValue, 0)

	for i := range source {
		if source[i].Value == "unavailable" || source[i].Value == "unknown" {
			continue
		}
		val, err := strconv.ParseFloat(source[i].Value, 64)
		if err != nil {
			return nil, err
		}
		res = append(res, NumericHistoryValue{
			Timestamp: source[i].Timestamp,
			Value:     val,
		})
	}

	return res, nil
}

var insideTemperatureSensorName = "sensor.living_room_sensor_temperature"
var insideHumiditySensorName = "sensor.living_room_sensor_humidity"
var outsideTemperatureSensorName = "sensor.aht10_temperature"
var outsideHumiditySensorName = "sensor.aht10_humidity"
var pressureSensorName = "sensor.living_room_sensor_pressure"
var hPaToMmHgCoeff = 1.33
var normalPressureMmHg = 760.0

func GetInsideTemperatureHumidity() (*TemperatureHumidityData, error) {
	return getTemperatureHumidity("Inside", insideTemperatureSensorName, insideHumiditySensorName)
}

func GetOutsideTemperatureHumidity() (*TemperatureHumidityData, error) {
	return getTemperatureHumidity("Outside", outsideTemperatureSensorName, outsideHumiditySensorName)
}

func GetPressure() (*PressureData, error) {
	pressure, err := DownloadSensorValueFromHA(pressureSensorName)
	if err != nil {
		return nil, err
	}
	pressureVal, err := strconv.ParseFloat(pressure, 64)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	pressureVal /= hPaToMmHgCoeff
	pressureHistory, err := DownloadSensorHistoryFromHA(pressureSensorName, now.Add(-60*time.Minute), now, false)
	if err != nil {
		return nil, err
	}
	pressureNumericSeries, err := ConvertToNumericSeries(pressureHistory)
	if err != nil {
		return nil, err
	}
	pressureSlope := Slope(pressureNumericSeries) * 3600 / hPaToMmHgCoeff
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
		PressureFalling:   pressureFalling,
		PressureSteady:    pressureSteady,
		PressureAboveNorm: pressureAboveNorm,
		PressureBelowNorm: !pressureAboveNorm,
		PressureDeltaInt:  pressureDeltaInt,
		PressureDeltaFrac: formatFrac(math.Abs(pressureDelta)),
	}, nil
}

func getTemperatureHumidity(title, temperatureSensorName, humiditySensorName string) (*TemperatureHumidityData, error) {
	temp, err := DownloadSensorValueFromHA(temperatureSensorName)
	if err != nil {
		return nil, err
	}
	tempVal, err := strconv.ParseFloat(temp, 64)
	if err != nil {
		return nil, err
	}
	humidity, err := DownloadSensorValueFromHA(humiditySensorName)
	if err != nil {
		return nil, err
	}
	humidityVal, err := strconv.ParseFloat(humidity, 64)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	tempHistory, err := DownloadSensorHistoryFromHA(temperatureSensorName, now.Add(-30*time.Minute), now, false)
	if err != nil {
		return nil, err
	}
	humidityHistory, err := DownloadSensorHistoryFromHA(humiditySensorName, now.Add(-30*time.Minute), now, false)
	if err != nil {
		return nil, err
	}
	tempNumericSeries, err := ConvertToNumericSeries(tempHistory)
	if err != nil {
		return nil, err
	}
	humidityNumericSeries, err := ConvertToNumericSeries(humidityHistory)
	if err != nil {
		return nil, err
	}
	tempSlope := Slope(tempNumericSeries) * 3600
	humiditySlope := Slope(humidityNumericSeries) * 3600
	tempRising := tempSlope >= 0.3
	tempFalling := tempSlope <= -0.3
	tempSteady := !(tempRising || tempFalling)
	humidityRising := humiditySlope >= 0.3
	humidityFalling := humiditySlope <= -0.3
	humiditySteady := !(humidityRising || humidityFalling)
	hundredPercentHumidity := humidityVal > 99.9
	return &TemperatureHumidityData{
		Title:              title,
		Warning:            false,
		TemperatureInt:     formatInt(tempVal),
		TemperatureFrac:    formatFrac(tempVal),
		TemperatureRising:  tempRising,
		TemperatureFalling: tempFalling,
		TemperatureSteady:  tempSteady,
		HumidityInt:        formatInt(humidityVal),
		HumidityFrac:       formatFrac(humidityVal),
		HumidityRising:     humidityRising,
		HumidityFalling:    humidityFalling,
		HumiditySteady:     humiditySteady,
		HundredPercentHumidity: hundredPercentHumidity,
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
	return strconv.Itoa(firstFracDigit)
}

// Calculates a non-vertical slope coefficient for a given set of points.
// It uses liner regression formula described here:
// https://en.wikipedia.org/wiki/Simple_linear_regression#Fitting_the_regression_line
// the result is the approximate  change of the value per second calculated over the data provided
// you may want to multiply by 3600 to get the hourly change
func Slope(res []NumericHistoryValue) float64 {
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
