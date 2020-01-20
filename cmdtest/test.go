package main

import (
	"net/http"
	"time"

	"github.com/golang/glog"
	"github.com/hydro-monitor/node/pkg/server"
)

const (
	postNodeMeasurementUrl = "http://antiguos.fi.uba.ar:443/api/nodes/%s/readings"
	NODE_NAME              = "1"
)

var client = &http.Client{
	Timeout: 10 * time.Second,
}

/*
func makeReq(uri string, m server.APIMeasurement) (*http.Response, error) {
	picturePath := m.Picture

	file, err := os.Open(picturePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	if err := writer.WriteField("timestamp", m.Time.Format(time.RFC3339)); err != nil {
		return nil, err
	}
	if err := writer.WriteField("waterLevel", fmt.Sprintf("%f", m.WaterLevel)); err != nil {
		return nil, err
	}

	part, err := writer.CreateFormFile("picture", filepath.Base(picturePath))
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(part, file); err != nil {
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}

	contentType := writer.FormDataContentType()
	res, err := http.Post(fmt.Sprintf(uri, NODE_NAME), contentType, body)
	if err != nil {
		return nil, err
	}
	return res, err
}

func main() {
	m := server.APIMeasurement{
		Time:       time.Now(),
		WaterLevel: 600,
		Picture:    "/Users/abarbetta/Desktop/lala.jpg",
	}

	res, err := makeReq(postNodeMeasurementUrl, m)
	if err != nil {
		glog.Errorf("Error making req: %v", err)
		return
	}
	defer res.Body.Close()

	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		glog.Fatal(err)
		return
	}
	bodyString := string(bodyBytes)
	glog.Infof("Status code: %d. Body: %v", res.StatusCode, bodyString)
	return
}
*/

func main() {
	m := server.APIMeasurement{
		Time:       time.Now(),
		WaterLevel: 600,
		Picture:    "/Users/abarbetta/Desktop/lala.jpg",
	}

	err := server.PostNodeMeasurement(m)
	if err != nil {
		glog.Errorf("Error making req: %v", err)
		return
	}
	return
}
