package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const (
	getNodeConfigurationUrl        = "https://my-json-server.typicode.com/hydro-monitor/web-api-mock/configurations/%s" // TODO Turn consts into env variables
	postNodeMeasurementUrl         = "http://antiguos.fi.uba.ar:443/api/nodes/%s/readings"
	getManualMeasurementRequestUrl = "https://my-json-server.typicode.com/hydro-monitor/web-api-mock/node/%s/requests"
	NODE_NAME                      = "1"
)

var client = &http.Client{Timeout: 10 * time.Second}

// estados(ID nodo (text),
//         nombre (text),
//         cantidad de fotos a tomar por medición (int),
//         cada cuantos ms tiempo toma medición (int),
//         límite de nivel de agua para pasar al estado anterior (float),
//         límite de nivel de agua para pasar al estado siguiente (float),
//         nombre estado anterior (text),
//         nombre estado siguiente (text))
type State struct {
	Name        string
	Interval    int
	UpperLimit  float64
	LowerLimit  float64
	PicturesNum int
	Next        string // State name (key)
	Prev        string // State name (key)
}

type APIConfigutation struct {
	States map[string]State `json:"states"`
}

type APIMeasurement struct {
	Time       time.Time `json:"timestamp"`
	WaterLevel float64   `json:"waterLevel"`
	Picture    string    `json:"picture"`
	WasManual bool `json:"wasManual"`
}

type APIMeasurementRequest struct {
	State string `json:"state"`
}

func GetNodeConfiguration() (*APIConfigutation, error) {
	resp, err := client.Get(fmt.Sprintf(getNodeConfigurationUrl, NODE_NAME))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respConfig := APIConfigutation{}
	err = json.NewDecoder(resp.Body).Decode(&respConfig)
	return &respConfig, err
}

func pictureUploadRequest(uri string, params map[string]string, path string) (*http.Request, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("Picture", filepath.Base(path))
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, file)

	for key, val := range params {
		_ = writer.WriteField(key, val)
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf(uri, NODE_NAME), body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, err
}

func PostNodeMeasurement(measurement APIMeasurement) error {
	extraParams := map[string]string {
		"Time": measurement.Time.String(),
		"WaterLevel": fmt.Sprintf("%f", measurement.WaterLevel),
	}
	request, err := pictureUploadRequest(postNodeMeasurementUrl, extraParams, measurement.Picture)
	if err != nil {
		return err
	}
	resp, err := client.Do(request)
	if err != nil {
		return err
	} 
	defer resp.Body.Close()
	return nil
}

// TODO Check if request with state is needed or the fact that a request itself exists
// is enough to know a manual measurement was requested.
// Also, we need another method to DELETE/PUT the manual request and let the server now the measurement was taken
func GetManualMeasurementRequest() (bool, error) {
	resp, err := client.Get(fmt.Sprintf(getManualMeasurementRequestUrl, NODE_NAME))
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	respMeasurementReq := APIMeasurementRequest{}
	if err := json.NewDecoder(resp.Body).Decode(&respMeasurementReq); err != nil {
		return false, err
	}
	if respMeasurementReq.State == "Pending" {
		return true, nil
	}
	return false, nil
}
