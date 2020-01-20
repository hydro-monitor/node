package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/golang/glog"
	"github.com/hydro-monitor/node/pkg/server"
)

const (
	postNodeMeasurementUrl = "http://192.168.1.12:8080/api/nodes/%s/readings"
	NODE_NAME              = "1"
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
	part, err := writer.CreateFormFile("Picture", filepath.Base(path))
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(part, file); err != nil {
		return nil, err
	}

	for key, val := range params {
		_ = writer.WriteField(key, val)
	}
	//
	if err := writer.WriteField("time", time.Now().String()); err != nil {
		return nil, err
	}
	if err := writer.WriteField("waterLevel", fmt.Sprintf("%f", 3.14)); err != nil {
		return nil, err
	}
	//

	if err := writer.Close(); err != nil {
		return nil, err
	}

	res, err := http.Post(fmt.Sprintf(uri, NODE_NAME), http.DetectContentType(body.Bytes()), body)
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
		Picture:    "/Users/abarbetta/Desktop/lala.jpg",
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
