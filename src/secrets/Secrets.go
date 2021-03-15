package secrets

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

const configFileName = "config.json"

type homeAssistantSettings struct {
	ServerProtocol string `json:"server_protocol"`
	ServerAddress  string `json:"server_address"`
	ServerPort     uint16 `json:"server_port"`
	Token          string `json:"token"`
}

type openWeatherMapSettings struct {
	ApiKey      string `json:"api_key"`
	PostCode    string `json:"post_code"`
	CountryCode string `json:"country_code"`
}

type configData struct {
	HomeAssistant  homeAssistantSettings  `json:"home_assistant"`
	OpenWeatherMap openWeatherMapSettings `json:"open_weather_map"`
}

var config *configData

func readConfig() {
	if config != nil {
		return
	}
	config = &configData{}
	buf, err := ioutil.ReadFile(configFileName)
	if err != nil {
		panic(fmt.Errorf("error reading config file: %v\n", err))
	}
	err = json.Unmarshal(buf, config)
	if err != nil {
		panic(fmt.Errorf("couldn't parse json file: %v\n", err))
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
