package main

import (
	"bytes"
	"fmt"
	"github.com/golang/glog"
	"github.com/hydro-monitor/node/pkg/server"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const (
	postNodeMeasurementUrl = "http://localhost:8080/api/nodes/%s/readings"
	NODE_NAME              = "lujan-1"
)

var client = &http.Client{
	Timeout: 10 * time.Second,
}

func makeReq(uri string, params map[string]string, path string) (*http.Response, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	if err := writer.WriteField("timestamp", time.Now().Format(time.RFC3339)); err != nil {
		return nil, err
	}
	if err := writer.WriteField("waterLevel", fmt.Sprintf("%f", 3.14)); err != nil {
		return nil, err
	}

	part, err := writer.CreateFormFile("picture", filepath.Base(path))
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(part, file); err != nil {
		return nil, err
	}

	/*for key, val := range params {
		_ = writer.WriteField(key, val)
	}*/

	if err := writer.Close(); err != nil {
		return nil, err
	}
    contentType := writer.FormDataContentType()
    glog.Info(contentType)
	res, err := http.Post(fmt.Sprintf(uri, NODE_NAME), contentType, body)
	if err != nil {
		return nil, err
	}
	//req.Header.Set("Content-Type", writer.FormDataContentType())
	return res, err
}

func main() {
	m := server.APIMeasurement{
		Time:       time.Now(),
		WaterLevel: 600,
		Picture:    "/Users/mporto/Desktop/maxresdefault.jpg",
	}

	res, err := makeReq(postNodeMeasurementUrl, nil, m.Picture)
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

	/*
		resp, err := client.Do(request)
		if err != nil {
			glog.Errorf("Error making request: %v", err)
			return
		}
		defer resp.Body.Close()
	*/
	return
}
