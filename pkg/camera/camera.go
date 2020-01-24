package camera

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dhowden/raspicam"
	"github.com/golang/glog"
)

const (
	picturesDir = "/home/pi/Documents/pictures"
)

type Camera struct{}

func (c *Camera) getStillConfig() *raspicam.Still {
	stillConfig := raspicam.NewStill()
	stillConfig.Quality = 20
	//stillConfig.Height = TODO set size
	//stillConfig.Width =
	stillConfig.Preview.Mode = raspicam.PreviewDisabled
	stillConfig.Timeout = time.Duration(500 * time.Millisecond)
	return stillConfig
}

func (c *Camera) TakeStill(stillName string) (string, error) {
	fileName := fmt.Sprintf("%s/%s", picturesDir, stillName)
	file, err := os.Create(fileName)
	if err != nil {
		glog.Errorf("Error creating file for picture: %v", err)
		return "", err
	}
	defer file.Close()

	stillConfig := c.getStillConfig()

	errCh := make(chan error)
	var errStr []string
	go func() {
		for x := range errCh {
			glog.Errorf("%v\n", x)
			errStr = append(errStr, x.Error())
		}
	}()

	glog.Info("Capturing still with picamera")
	raspicam.Capture(stillConfig, file, errCh)

	if len(errStr) > 0 {
		return fileName, fmt.Errorf(strings.Join(errStr, "\n"))
	}
	return fileName, nil
}
