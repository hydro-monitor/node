package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gocql/gocql"
	"github.com/golang/glog"
	"github.com/hashicorp/go-retryablehttp"

	"github.com/hydro-monitor/node/pkg/envconfig"
)

const (
	contentTypeHeader = "Content-Type"
	contentTypeHeaderValue = "application/json"
	authorizationHeader = "Authorization"
	authorizationHeaderValue = "Bearer %s"
)

// Server represents a server for a specific node
type Server struct {
	client                         *http.Client
	notFoundClient                 *http.Client
	nodeName                       string
	nodePassword                   string
	secret                         string
	getNodeConfigurationURL        string
	postNodeMeasurementURL         string
	postNodePictureURL             string
	getManualMeasurementRequestURL string
}

// NewServer creates and returns a server taking nodeName and urls from env config
func NewServer() *Server {
	env := envconfig.New()

	retryClient := retryablehttp.NewClient()
	retryClient.RetryWaitMin = time.Duration(env.HTTPClientRetryWaitMin) * time.Second
	retryClient.RetryMax = env.HTTPClientMaxRetries
	client := retryClient.StandardClient()

	retryOnNotFoundClient := retryablehttp.NewClient()
	retryOnNotFoundClient.RetryWaitMin = time.Duration(env.HTTPClientRetryWaitMin) * time.Second
	retryOnNotFoundClient.RetryMax = env.HTTPClientMaxRetries
	retryOnNotFoundClient.CheckRetry = retryOnNotFoundPolicy
	notFoundClient := retryOnNotFoundClient.StandardClient()

	return &Server{
		client:                         client,
		notFoundClient:                 notFoundClient,
		nodeName:                       env.NodeName,
		nodePassword:                   env.NodePassword,
		secret:                         env.Secret,
		getNodeConfigurationURL:        env.GetNodeConfigurationURL,
		postNodeMeasurementURL:         env.PostNodeMeasurementURL,
		postNodePictureURL:             env.PostNodePictureURL,
		getManualMeasurementRequestURL: env.GetManualMeasurementRequestURL,
	}
}

// State represents a state of a configuration
type State struct {
	// Time interval between measurements in seconds
	// Intervalo de tiempo entre mediciones en segundos
	Interval    int
	// Minimum water level limit to be in this state
	// Límite de nivel de agua para pasar al estado anterior
	UpperLimit  float64
	// Maximum water level to be in this state
	// Límite de nivel de agua para pasar al estado siguiente
	LowerLimit  float64
	// Amount of pictures taken per measurement
	// Cantidad de fotos tomadas por medición
	PicturesNum int
}

// APIConfigutation represents a node configuration response from the hydro monitor server
type APIConfigutation struct {
	States map[string]State `json:"states,inline"`
}

// APIMeasurement represents a measurement creation request for the hydro monitor server
type APIMeasurement struct {
	Time          time.Time `json:"timestamp"`
	WaterLevel    float64   `json:"waterLevel"`
	ManualReading bool      `json:"manualReading"`
}

// APIMeasurementResponse represents a measurement creation response from the hydro monitor server
type APIMeasurementResponse struct {
	APIMeasurement `json:",inline"`
	ReadingID      gocql.UUID `json:"readingId"`
}

// APIPicture represents a picture creation request for the hydro monitor server
type APIPicture struct {
	MeasurementID gocql.UUID `json:"measurementId"`
	Picture       string     `json:"picture"`
	PictureNumber int        `json:"pictureNumber"`
}

// APIMeasurementRequest represents a manual measurement request response from the hydro monitor server
type APIMeasurementRequest struct {
	ManualReading bool `json:"manualReading"`
}

// generateJWT generates a json web token based on node password and secret
func generateJWT(nodePassword, secret string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["exp"] = time.Now().Add(time.Hour * 8).Unix()
	claims["password"] = nodePassword

	t, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}
	glog.Infof("JWT generated: %s", t)
	return t, nil
}

// doPostRequest generates JWT, sets headers and performs POST request
func (s *Server) doPostRequest(url, contentType string, body io.Reader, retryOnNotFound bool) (*http.Response, error) {
	token, err := generateJWT(s.nodePassword, s.secret)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set(contentTypeHeader, contentType)
	req.Header.Set(authorizationHeader, fmt.Sprintf(authorizationHeaderValue, token))
	if retryOnNotFound {
		return s.notFoundClient.Do(req)
	}
	return s.client.Do(req)
}

// doGetRequest generates JWT, sets headers and performs GET request
func (s *Server) doGetRequest(url string, retryOnNotFound bool) (*http.Response, error) {
	token, err := generateJWT(s.nodePassword, s.secret)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set(authorizationHeader, fmt.Sprintf(authorizationHeaderValue, token))
	if retryOnNotFound {
		return s.notFoundClient.Do(req)
	}
	return s.client.Do(req)
}

// requestIsSuccessful returns true if the status code is 2XX, false otherwise
func requestIsSuccessful(statusCode int) bool {
	return (statusCode >= 200 && statusCode <= 299)
}

// GetNodeConfiguration returns node configuration from hydro monitor server
func (s *Server) GetNodeConfiguration() (*APIConfigutation, error) {
	resp, err := s.doGetRequest(fmt.Sprintf(s.getNodeConfigurationURL, s.nodeName), false)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("Node has no configuration loaded")
	}
	if !requestIsSuccessful(resp.StatusCode) {
		return nil, fmt.Errorf("Error requesting node configuration. code: %v", resp.StatusCode)
	}

	statesMap := make(map[string]State)
	err = json.NewDecoder(resp.Body).Decode(&statesMap)
	respConfig := APIConfigutation{
		States: statesMap,
	}
	return &respConfig, err
}

// PostNodeMeasurement sends new measurement to hydro monitor server
func (s *Server) PostNodeMeasurement(measurement APIMeasurement) (*gocql.UUID, error) {
	requestByte, err := json.Marshal(measurement)
	if err != nil {
		return nil, err
	}
	requestReader := bytes.NewReader(requestByte)
	
	res, err := s.doPostRequest(fmt.Sprintf(s.postNodeMeasurementURL, s.nodeName), contentTypeHeaderValue, requestReader, false)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if !requestIsSuccessful(res.StatusCode) {
		return nil, fmt.Errorf("Error creating node measurement. code: %v", res.StatusCode)
	}

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

// PostNodePicture sends new measurement picture to hydro monitor server
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
	res, err := s.doPostRequest(fmt.Sprintf(s.postNodePictureURL, s.nodeName, measurementID), contentType, body, true)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if !requestIsSuccessful(res.StatusCode) {
		return fmt.Errorf("Error creating node measurement photo. code: %v", res.StatusCode)
	}

	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		glog.Errorf("Error reading response body for picture upload: %v", err)
		return err
	}
	bodyString := string(bodyBytes)
	glog.Infof("Status code for picture upload: %d. Body: %v", res.StatusCode, bodyString)

	return nil
}

// GetManualMeasurementRequest returns true if manual measurement is requested
func (s *Server) GetManualMeasurementRequest() (bool, error) {
	resp, err := s.doGetRequest(fmt.Sprintf(s.getManualMeasurementRequestURL, s.nodeName), false)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if !requestIsSuccessful(resp.StatusCode) {
		return false, fmt.Errorf("Error getting manual measurement request. code: %v", resp.StatusCode)
	}

	respMeasurementReq := APIMeasurementRequest{}
	if err := json.NewDecoder(resp.Body).Decode(&respMeasurementReq); err != nil {
		return false, err
	}
	if respMeasurementReq.ManualReading {
		return true, nil
	}
	return false, nil
}

// retryOnNotFoundPolicy is the same as DefaultRetryPolicy, except it
// retries if resp status code was Not Found (404)
func retryOnNotFoundPolicy(ctx context.Context, resp *http.Response, err error) (bool, error) {
	retry, reqErr := retryablehttp.DefaultRetryPolicy(ctx, resp, err);
	if !retry && resp.StatusCode == 404 {
		// Retry in case of 404 just in case of consistency between servers is not met yet
		glog.Infof("Response (%v) was Not Found, retrying", resp)
		return true, reqErr
	}
	return retry, reqErr
}