package config

import (
	"encoding/json"
	"github.com/rotisserie/eris"
	"os"
	"path"
)

const configFileName = "config.json"

func GetRootDir() string {
	exec := os.Args[0]
	dir := path.Dir(exec)
	if dir == "." {
		curDir, err := os.Getwd()
		if err != nil {
			panic(eris.ToString(eris.Wrap(err, "Error getting current dir"), true))
		}
		dir = curDir
	}
	return dir
}

type homeAssistantSettings struct {
	ServerProtocol            string `json:"server_protocol"`
	ServerAddress             string `json:"server_address"`
	ServerPort                uint16 `json:"server_port"`
	Token                     string `json:"token"`
	InternalTemperatureSensor string `json:"internal_temperature_sensor"`
	ExternalTemperatureSensor string `json:"external_temperature_sensor"`
	InternalHumiditySensor    string `json:"internal_humidity_sensor"`
	ExternalHumiditySensor    string `json:"external_humidity_sensor"`
	PressureSensor            string `json:"pressure_sensor"`
}

type openWeatherMapSettings struct {
	ApiKey      string `json:"api_key"`
	PostCode    string `json:"post_code"`
	CountryCode string `json:"country_code"`
}

type daylightSettings struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type configData struct {
	HomeAssistant    homeAssistantSettings   `json:"home_assistant"`
	OpenWeatherMap   openWeatherMapSettings  `json:"open_weather_map"`
	SpecialDays      []*SpecialDayOrInterval `json:"special_days"`
	DaylightSettings daylightSettings        `json:"daylight_settings"`
}

type SpecialDayOrInterval struct {
	Index           int    `json:"index"`
	Id              string `json:"id"`
	DisplayText     string `json:"display_text"`
	Type            string `json:"type"`
	StartDateDay    int    `json:"start_date_day"`
	StartDateMonth  int    `json:"start_date_month"`
	StartDateYear   int    `json:"start_date_year"`
	EndDateDay      int    `json:"end_date_day"`
	EndDateMonth    int    `json:"end_date_month"`
	EndDateYear     int    `json:"end_date_year"`
	IsPublicHoliday bool   `json:"is_public_holiday"`
	IsSchoolHoliday bool   `json:"is_school_holiday"`
}

func readConfig() (*configData, error) {
	fileName := path.Join(GetRootDir(), configFileName)
	config := configData{}
	buf, err := os.ReadFile(fileName)
	if err != nil {
		return nil, eris.Wrap(err, "error reading config file")
	}
	err = json.Unmarshal(buf, &config)
	if err != nil {
		return nil, eris.Wrap(err, "couldn't parse json file")
	}
	return &config, nil
}

func (c *configApi) saveConfig() {
	fileName := path.Join(GetRootDir(), configFileName)
	data, err := json.Marshal(c.config)
	if err != nil {
		panic(eris.ToString(eris.Wrap(err, "error serializing config"), true))
	}
	err = os.WriteFile(fileName, data, 0644)
	if err != nil {
		panic(eris.ToString(eris.Wrap(err, "error writing to config file"), true))
	}
}

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
	SetSpecialDays(specialDays []*SpecialDayOrInterval)
	GetSpecialDays() []*SpecialDayOrInterval
	GetDaylightCoordinates() (float64, float64)
	GetSimpleRefresh() bool
	ResetSimpleRefresh()
	SetSimpleRefresh()
	GetHAToken() string
	GetHAProtocol() string
	GetHAHost() string
	GetHAPort() uint16
	GetOpenWeatherMapAPIKey() string
	GetOpenWeatherMapPostCode() string
	GetOpenWeatherMapCountryCode() string
	GetCalendarRedraw() bool
	ResetCalendarRedraw()
	SetCalendarRedraw()
	GetRedrawAll() bool
	ResetRedrawAll()
	SetRedrawAll()
}

type configApi struct {
	config         *configData
	simpleRefresh  bool
	calendarRedraw bool
	redrawAll      bool
}

func (c *configApi) SetCalendarRedraw() {
	c.calendarRedraw = true
}

func (c *configApi) SetRedrawAll() {
	c.redrawAll = true
}

func (c *configApi) GetCalendarRedraw() bool {
	return c.calendarRedraw
}

func (c *configApi) ResetCalendarRedraw() {
	c.calendarRedraw = false
}

func (c *configApi) GetRedrawAll() bool {
	return c.redrawAll
}

func (c *configApi) ResetRedrawAll() {
	c.redrawAll = true
}

const picnicPointLatitude = -33.969526
const picnicPointLongitude = 150.998711
const moscowLatitude = 55.643940
const moscowLongitude = 37.528860

func (c *configApi) GetDaylightCoordinates() (float64, float64) {
	return c.config.DaylightSettings.Latitude, c.config.DaylightSettings.Longitude
}

func (c *configApi) GetSimpleRefresh() bool {
	return c.simpleRefresh
}

func (c *configApi) ResetSimpleRefresh() {
	c.simpleRefresh = false
}

func (c *configApi) SetSimpleRefresh() {
	c.simpleRefresh = true
}

func (c *configApi) GetHAToken() string {
	return c.config.HomeAssistant.Token
}

func (c *configApi) GetHAProtocol() string {
	return c.config.HomeAssistant.ServerProtocol
}

func (c *configApi) GetHAHost() string {
	return c.config.HomeAssistant.ServerAddress
}

func (c *configApi) GetHAPort() uint16 {
	return c.config.HomeAssistant.ServerPort
}

func (c *configApi) GetOpenWeatherMapAPIKey() string {
	return c.config.OpenWeatherMap.ApiKey
}

func (c *configApi) GetOpenWeatherMapPostCode() string {
	return c.config.OpenWeatherMap.PostCode
}

func (c *configApi) GetOpenWeatherMapCountryCode() string {
	return c.config.OpenWeatherMap.CountryCode
}

func (c *configApi) SetSpecialDays(specialDays []*SpecialDayOrInterval) {
	c.config.SpecialDays = specialDays
	c.calendarRedraw = true
	c.saveConfig()
}

func (c *configApi) GetSpecialDays() []*SpecialDayOrInterval {
	if c.config.SpecialDays == nil {
		return []*SpecialDayOrInterval{}
	}
	return c.config.SpecialDays
}

func (c *configApi) GetInternalTemperatureSensorName() string {
	return c.config.HomeAssistant.InternalTemperatureSensor
}

func (c *configApi) GetInternalHumiditySensorName() string {
	return c.config.HomeAssistant.InternalHumiditySensor
}

func (c *configApi) GetExternalTemperatureSensorName() string {
	return c.config.HomeAssistant.ExternalTemperatureSensor
}

func (c *configApi) GetExternalHumiditySensorName() string {
	return c.config.HomeAssistant.ExternalHumiditySensor
}

func (c *configApi) GetPressureSensorName() string {
	return c.config.HomeAssistant.PressureSensor
}

func (c *configApi) RedrawAll() {
	c.redrawAll = true
}

func (c *configApi) SetInternalTemperatureSensorName(sensorName string) {
	c.config.HomeAssistant.InternalTemperatureSensor = sensorName
	c.saveConfig()
}

func (c *configApi) SetInternalHumiditySensorName(sensorName string) {
	c.config.HomeAssistant.InternalHumiditySensor = sensorName
	c.saveConfig()
}

func (c *configApi) SetExternalTemperatureSensorName(sensorName string) {
	c.config.HomeAssistant.ExternalTemperatureSensor = sensorName
	c.saveConfig()
}

func (c *configApi) SetExternalHumiditySensorName(sensorName string) {
	c.config.HomeAssistant.ExternalHumiditySensor = sensorName
	c.saveConfig()
}

func (c *configApi) SetPressureSensorName(sensorName string) {
	c.config.HomeAssistant.PressureSensor = sensorName
	c.saveConfig()
}

func NewConfigApi() (ConfigApi, error) {
	config, err := readConfig()
	if err != nil {
		return nil, eris.Wrap(err, "Error reading config file")
	}
	return &configApi{
		config:         config,
		simpleRefresh:  false,
		calendarRedraw: false,
		redrawAll:      false,
	}, nil
}
