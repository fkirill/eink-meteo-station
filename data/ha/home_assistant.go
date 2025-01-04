package ha

import (
	"encoding/json"
	"errors"
	"fkirill.org/eink-meteo-station/config"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type HomeAssistantHistoryItem struct {
	Timestamp time.Time `json:"last_changed"`
	Value     string    `json:"state"`
}

type NumericHistoryValue struct {
	Timestamp time.Time
	Value     float64
}

type HomeAssistantApi interface {
	DownloadSensorValueFromHA(sensorId string) (string, error)
	DownloadSensorHistoryFromHA(sensorId string, startTime, endTime time.Time, significantOnly bool) ([]*HomeAssistantHistoryItem, error)
}

type homeAssistantApi struct {
	config config.ConfigApi
}

func (h *homeAssistantApi) getBearerToken() string {
	return fmt.Sprintf("Bearer %s", h.config.GetHAToken())
}

func (h *homeAssistantApi) getHAProtocolHostPort() string {
	return fmt.Sprintf("%s://%s:%s", h.config.GetHAProtocol(), h.config.GetHAHost(), h.config.GetHAPort())
}

func (h homeAssistantApi) DownloadSensorValueFromHA(sensorId string) (string, error) {
	client := &http.Client{}
	url := fmt.Sprintf("%s/api/states/%s", h.getHAProtocolHostPort(), sensorId)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Authorization", h.getBearerToken())
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

func (h homeAssistantApi) DownloadSensorHistoryFromHA(sensorId string, startTime, endTime time.Time, significantOnly bool) ([]*HomeAssistantHistoryItem, error) {
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
	req.Header.Add("Authorization", h.getBearerToken())
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
	var res [][]*HomeAssistantHistoryItem
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

func NewHomeAssistantApi(config config.ConfigApi) HomeAssistantApi {
	return &homeAssistantApi{config}
}
