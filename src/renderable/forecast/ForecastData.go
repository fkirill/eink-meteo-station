package forecast

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"secrets"
	"sort"
	"time"
)

var zipCode = secrets.GetOpenWeatherMapPostCode()
var countryCode = secrets.GetOpenWeatherMapCountryCode()
var apiKey = secrets.GetOpenWeatherMapAPIKey()

var queryURL = "https://api.openweathermap.org/data/2.5/forecast?zip=" + zipCode + "," + countryCode + "&appid=" + apiKey + "&units=metric"

// See https://openweathermap.org/forecast5#parameter for documentation

type latLon struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type weatherDataCity struct {
	Id         int    `json:"id"`
	Name       string `json:"name"`
	Coord      latLon `json:"coord"`
	Country    string `json:"country"`
	Timezone   int    `json:"timezone"`
	Sunrise    int64  `json:"sunrise"`
	Sunset     int64  `json:"sunset"`
	Population int64  `json:"population"`
}

type weatherDataItemSys struct {
	Pod string `json:"pod"` // partOfDay: n = night, d = day
}

type weatherDataItemMain struct {
	Temp      float64 `json:"temp"`
	FeelsLike float64 `json:"feels_like"`
	TempMin   float64 `json:"temp_min"`
	TempMax   float64 `json:"temp_max"`
	Pressure  float64 `json:"pressure"`
	SeaLevel  float64 `json:"sea_level"`  // pressure at sea level
	GrndLevel float64 `json:"grnd_level"` // pressure at ground level
	Humidity  float64 `json:"humidity"`
	TempKF    float64 `json:"temp_kf"` // internal
}

type weatherDataItemWeather struct {
	Id          int    `json:"id"`
	Main        string `json:"main"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

type weatherDataItemWind struct {
	Speed float64 `json:"speed"`
	Deg   float64 `json:"deg"`
}

type weatherDataItem struct {
	Dt         int64                    `json:"dt"`
	Main       weatherDataItemMain      `json:"main"`
	Weather    []weatherDataItemWeather `json:"weather"`
	Clouds     map[string]float64       `json:"clouds"`
	Wind       weatherDataItemWind      `json:"wind"`
	Visibility int                      `json:"visibility"`
	Pop        float64                  `json:"pop"`
	Rain       map[string]float64       `json:"rain"`
	Sys        weatherDataItemSys       `json:"sys"`
	Dt_Txt     string                   `json:"dt_txt"`
}

type weatherData struct {
	Count int               `json:"cnt"`
	List  []weatherDataItem `json:"list"`
	City  weatherDataCity   `json:"city"`
}

type ForecastDataDay struct {
	EpochDay             int
	Date                 time.Time
	MinTemp              float64
	MaxTemp              float64
	MaxChanceOfRain      float64
	ExpectedRainAmountMm float64
	MaxWindKmh           float64
	WeatherType          int
}

type ForecastDataGraph struct {
	DateTime     time.Time
	Temperature  float64
	Humidity     float64
	Clouds       float64
	ChanceOfRain float64
	WindKmh      float64
}

type ForecastDataDaySlice []ForecastDataDay
type ForecastDataGraphSlice []ForecastDataGraph

func (s ForecastDataDaySlice) Len() int           { return len(s) }
func (s ForecastDataDaySlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s ForecastDataDaySlice) Less(i, j int) bool { return s[i].EpochDay < s[j].EpochDay }

func (s ForecastDataGraphSlice) Len() int           { return len(s) }
func (s ForecastDataGraphSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s ForecastDataGraphSlice) Less(i, j int) bool { return s[i].DateTime.Before(s[j].DateTime) }

type ForecastData struct {
	Days      []ForecastDataDay
	GraphData []ForecastDataGraph
}

func GetWeatherData() (*ForecastData, error) {
	response, err := http.Get(queryURL)
	if err != nil {
		urlErr := err.(*url.Error)
		if urlErr.Timeout() || urlErr.Temporary() {
			// can retry later, returning empty structure and no error
			return &ForecastData{}, nil
		}
		return nil, err
	}
	defer response.Body.Close()
	weather := weatherData{}
	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(responseData, &weather)
	if err != nil {
		return nil, err
	}
	forecastData, err := transformIntoForecastData(weather)
	return forecastData, nil
}

func transformIntoForecastData(weather weatherData) (*ForecastData, error) {
	daysMap := make(map[int]*ForecastDataDay)
	graphMap := make(map[int]*ForecastDataGraph)
	for _, item := range weather.List {
		dttm := time.Unix(item.Dt, 0)
		epochDay := int(item.Dt / 86400)
		curDay, exists := daysMap[epochDay]
		if !exists {
			curDay = &ForecastDataDay{
				EpochDay:             epochDay,
				Date:                 dttm.Truncate(24 * time.Hour),
				MinTemp:              200,
				MaxTemp:              -200,
				MaxChanceOfRain:      0,
				ExpectedRainAmountMm: 0,
				MaxWindKmh:           -200,
				WeatherType:          0,
			}
			daysMap[epochDay] = curDay
		}
		clouds, _ := item.Clouds["all"]
		graphMap[int(item.Dt)] = &ForecastDataGraph{
			DateTime:     dttm,
			Temperature:  item.Main.Temp,
			Humidity:     item.Main.Humidity,
			Clouds:       clouds,
			ChanceOfRain: 0,
			WindKmh:      item.Wind.Speed,
		}
		if curDay.MinTemp > item.Main.Temp {
			curDay.MinTemp = item.Main.Temp
		}
		if curDay.MaxTemp < item.Main.Temp {
			curDay.MaxTemp = item.Main.Temp
		}
		curDay.ExpectedRainAmountMm += item.Rain["3h"]
		if curDay.MaxWindKmh < item.Wind.Speed {
			curDay.MaxWindKmh = item.Wind.Speed
		}
	}
	days := make([]ForecastDataDay, 0)
	for _, v := range daysMap {
		days = append(days, *v)
	}
	graphData := make([]ForecastDataGraph, 0)
	for _, v := range graphMap {
		graphData = append(graphData, *v)
	}
	sort.Sort(ForecastDataDaySlice(days))
	sort.Sort(ForecastDataGraphSlice(graphData))
	return &ForecastData{
		Days:      days,
		GraphData: graphData,
	}, nil
}
