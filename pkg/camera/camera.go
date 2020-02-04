package camera

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/dhowden/raspicam"
	"github.com/golang/glog"

	"github.com/hydro-monitor/node/pkg/envconfig"
)

type Camera struct {
	sync.Mutex
	picturesDir string
}

func NewCamera() *Camera {
	env := envconfig.New()
	return &Camera{
		picturesDir: env.PicturesDir,
	}
}

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
	c.Lock()
	defer c.Unlock()
	fileName := fmt.Sprintf("%s/%s", c.picturesDir, stillName)
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
