package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gocql/gocql"
	"github.com/golang/glog"

	"github.com/hydro-monitor/node/pkg/envconfig"
)

type Server struct {
	client                         *http.Client
	nodeName                       string
	getNodeConfigurationURL        string
	postNodeMeasurementURL         string
	postNodePictureURL             string
	getManualMeasurementRequestURL string
}

func NewServer() *Server {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	env := envconfig.New()

	return &Server{
		client:                         client,
		nodeName:                       env.NodeName,
		getNodeConfigurationURL:        env.GetNodeConfigurationURL,
		postNodeMeasurementURL:         env.PostNodeMeasurementURL,
		postNodePictureURL:             env.PostNodePictureURL,
		getManualMeasurementRequestURL: env.GetManualMeasurementRequestURL,
	}
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
	WasManual  bool      `json:"wasManual"`
}

type APIMeasurementResponse struct {
	APIMeasurement `json:",inline"`
	ReadingID      gocql.UUID `json:"readingId"`
}

type APIPicture struct {
	MeasurementID gocql.UUID `json:"measurementId"`
	Picture       string     `json:"picture"`
	PictureNumber int        `json:"pictureNumber"`
}

type APIMeasurementRequest struct {
	ManualReading bool `json:"manualReading"`
}

func (s *Server) GetNodeConfiguration() (*APIConfigutation, error) {
	url := fmt.Sprintf(s.getNodeConfigurationURL, s.nodeName)
	glog.Infof("url: %v", url) // FIXME remove
	resp, err := s.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respConfig := APIConfigutation{}
	err = json.NewDecoder(resp.Body).Decode(&respConfig)
	glog.Infof("respConfig: %v", respConfig) // FIXME remove
	return &respConfig, err
}

func (s *Server) PostNodeMeasurement(measurement APIMeasurement) (*gocql.UUID, error) {
	requestByte, _ := json.Marshal(measurement)
	requestReader := bytes.NewReader(requestByte)
	res, err := s.client.Post(fmt.Sprintf(s.postNodeMeasurementURL, s.nodeName), "application/json", requestReader)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		glog.Errorf("Error reading response body for measurement creation: %v", err)
		return nil, err
	}
	bodyString := string(bodyBytes)
	glog.Infof("Status code for measurement creation: %d. Body: %v", res.StatusCode, bodyString)

	var resObj APIMeasurementResponse
	if err := json.Unmarshal(bodyBytes, &resObj); err != nil {
		glog.Errorf("Error unmarshaling body %v", err)
		return nil, err
	}

	glog.Infof("Returning measurement ID: %v", &resObj.ReadingID)
	return &resObj.ReadingID, nil
}

func (s *Server) PostNodePicture(measurement APIPicture) error {
	measurementID := measurement.MeasurementID
	picturePath := measurement.Picture

	file, err := os.Open(picturePath)
	if err != nil {
		return err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	if err := writer.WriteField("pictureNumber", fmt.Sprintf("%d", measurement.PictureNumber)); err != nil {
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
	res, err := http.Post(fmt.Sprintf(s.postNodePictureURL, s.nodeName, measurementID), contentType, body)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		glog.Errorf("Error reading response body for picture upload: %v", err)
		return err
	}
	bodyString := string(bodyBytes)
	glog.Infof("Status code for picture upload: %d. Body: %v", res.StatusCode, bodyString)

	return nil
}

// TODO Check if request with state is needed or the fact that a request itself exists
// is enough to know a manual measurement was requested.
// Also, we need another method to DELETE/PUT the manual request and let the server now the measurement was taken
func (s *Server) GetManualMeasurementRequest() (bool, error) {
	resp, err := s.client.Get(fmt.Sprintf(s.getManualMeasurementRequestURL, s.nodeName))
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	respMeasurementReq := APIMeasurementRequest{}
	if err := json.NewDecoder(resp.Body).Decode(&respMeasurementReq); err != nil {
		return false, err
	}
	if respMeasurementReq.ManualReading {
		return true, nil
	}
	return false, nil
}
