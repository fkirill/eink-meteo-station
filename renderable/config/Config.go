package config

import (
	"encoding/json"
	"fmt"
	"os"
)

const configFileName = "config.json"

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

var config *configData

func readConfig() {
	if config != nil {
		return
	}
	config = &configData{}
	buf, err := os.ReadFile(configFileName)
	if err != nil {
		panic(fmt.Errorf("error reading config file: %v\n", err))
	}
	err = json.Unmarshal(buf, config)
	if err != nil {
		panic(fmt.Errorf("couldn't parse json file: %v\n", err))
	}
}

func saveConfig() {
	data, err := json.Marshal(&config)
	if err != nil {
		panic(fmt.Errorf("error serializing config: %v\n", err))
	}
	err = os.WriteFile(configFileName, data, 0777)
	if err != nil {
		panic(fmt.Errorf("error writitng to config file: %v\n", err))
	}
}

func init() {
	readConfig()
}

func GetHAToken() string {
	return config.HomeAssistant.Token
}

func GetHAProtocol() string {
	return config.HomeAssistant.ServerProtocol
}

func GetHAHost() string {
	return config.HomeAssistant.ServerAddress
}

func GetHAPort() uint16 {
	return config.HomeAssistant.ServerPort
}

func GetOpenWeatherMapAPIKey() string {
	return config.OpenWeatherMap.ApiKey
}

func GetOpenWeatherMapPostCode() string {
	return config.OpenWeatherMap.PostCode
}

func GetOpenWeatherMapCountryCode() string {
	return config.OpenWeatherMap.CountryCode
}

func GetInternalTemperatureSensor() string {
	return config.HomeAssistant.InternalTemperatureSensor
}

func GetExternalTemperatureSensor() string {
	return config.HomeAssistant.ExternalTemperatureSensor
}

func GetInternalHumiditySensor() string {
	return config.HomeAssistant.InternalHumiditySensor
}

func GetExternalHumiditySensor() string {
	return config.HomeAssistant.ExternalHumiditySensor
}

func GetPressureSensor() string {
	return config.HomeAssistant.PressureSensor
}

func SetInternalTemperatureSensor(sensorName string) {
	config.HomeAssistant.InternalTemperatureSensor = sensorName
	saveConfig()
}

func SetExternalTemperatureSensor(sensorName string) {
	config.HomeAssistant.ExternalTemperatureSensor = sensorName
	saveConfig()
}

func SetInternalHumiditySensor(sensorName string) {
	config.HomeAssistant.InternalHumiditySensor = sensorName
	saveConfig()
}

func SetExternalHumiditySensor(sensorName string) {
	config.HomeAssistant.ExternalHumiditySensor = sensorName
	saveConfig()
}

func SetPressureSensor(sensorName string) {
	config.HomeAssistant.PressureSensor = sensorName
	saveConfig()
}

func GetSpecialDays() []*SpecialDayOrInterval {
	if config.SpecialDays == nil {
		return []*SpecialDayOrInterval{}
	}
	return config.SpecialDays
}

func SetSpecialDays(specialDays []*SpecialDayOrInterval) {
	config.SpecialDays = specialDays
	saveConfig()
}

const picnicPointLatitude = -33.969526
const picnicPointLongitude = 150.998711
const moscowLatitude = 55.643940
const moscowLongitude = 37.528860

func GetDaylightCoordinates() (float64, float64) {
	return config.DaylightSettings.Latitude, config.DaylightSettings.Longitude
}

var simpleRefresh = false

func GetSimpleRefresh() bool {
	return simpleRefresh
}

func ResetSimpleRefresh() {
	simpleRefresh = false
}

func SetSimpleRefresh() {
	simpleRefresh = true
}
