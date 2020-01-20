package server

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/golang/glog"
)

const (
	getNodeConfigurationUrl        = "https://my-json-server.typicode.com/hydro-monitor/web-api-mock/configurations/%s" // TODO Turn consts into env variables
	postNodeMeasurementUrl         = "http://antiguos.fi.uba.ar:443/api/nodes/%s/readings"
	getManualMeasurementRequestUrl = "https://my-json-server.typicode.com/hydro-monitor/web-api-mock/node/%s/requests"
	NODE_NAME                      = "1"
)

//Configure TLS, etc.
var tr = &http.Transport{
	TLSClientConfig: &tls.Config{
		InsecureSkipVerify: true,
	},
}

var client = &http.Client{
	Timeout:   10 * time.Second,
	Transport: tr,
}

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
	//TODO Add when manual measurements are implemented WasManual bool `json:"wasManual"`
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

func PostNodeMeasurement(measurement APIMeasurement) error {
	picturePath := "/home/pi/Documents/pictures/lala" // FIXME Remove to send current image

	file, err := os.Open(picturePath)
	if err != nil {
		return err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	if err := writer.WriteField("timestamp", measurement.Time.Format(time.RFC3339)); err != nil {
		return err
	}
	if err := writer.WriteField("waterLevel", fmt.Sprintf("%f", measurement.WaterLevel)); err != nil {
		return err
	}

	part, err := writer.CreateFormFile("picture", filepath.Base(picturePath))
	if err != nil {
		return err
	}
	if _, err := io.Copy(part, file); err != nil {
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}

	contentType := writer.FormDataContentType()
	res, err := http.Post(fmt.Sprintf(postNodeMeasurementUrl, NODE_NAME), contentType, body)
	if err != nil {
		return err
	}

	/// DEBUG Block FIXME Remove
	defer res.Body.Close()

	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		glog.Errorf("DEBUG Error reading response body: %v", err)
		return err
	}
	bodyString := string(bodyBytes)
	glog.Infof("DEBUG Status code: %d. Body: %v", res.StatusCode, bodyString)
	///

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
