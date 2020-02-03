package envconfig

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
    NodeName string
    InitialTriggerInterval        int
	ConfigurationUpdateInterval   int
    ManualMeasurementPollInterval int
    PicturesDir string
    SerialPort string
    Baud int
    GetNodeConfigurationUrl string
    PostNodeMeasurementUrl string
    PostNodePictureUrl string
    GetManualMeasurementRequestUrl string
}

// New returns a new Config struct
func New() *Config {
	return &Config{
        NodeName: getEnv("NODE_NAME", "1"),
		InitialTriggerInterval: getEnvAsInt("INITIAL_TRIGGER_INTERVAL", 10),
		ConfigurationUpdateInterval: getEnvAsInt("CONFIGURATION_UPDATE_INTERVAL", 60),
        ManualMeasurementPollInterval:  getEnvAsInt("MANUAL_MEASUREMENT_POLL_INTERVAL", 180),
        PicturesDir: getEnv("PICTURES_DIR", "/home/pi/Documents/pictures"),
        SerialPort: getEnv("SERIAL_PORT", "/dev/ttyACM0"),
        Baud: getEnvAsInt("BAUD", 9600),
        GetNodeConfigurationUrl: getEnv("GET_NODE_CONFIGURATION_URL", "https://my-json-server.typicode.com/hydro-monitor/web-api-mock/configurations/%s"),
        PostNodeMeasurementUrl: getEnv("POST_NODE_MEASUREMENT_URL", "http://antiguos.fi.uba.ar:443/api/nodes/%s/readings"),
        PostNodePictureUrl: getEnv("POST_NODE_PICTURE_URL", "http://antiguos.fi.uba.ar:443/api/api/nodes/%s/readings/%s/photos"),
        GetManualMeasurementRequestUrl: getEnv("GET_MANUAL_MEASUREMENT_REQUEST_URL", "https://my-json-server.typicode.com/hydro-monitor/web-api-mock/requests/%s"),
	}
}

// Simple helper function to read an environment or return a default value
func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
}

// Simple helper function to read an environment variable into integer or return a default value
func getEnvAsInt(name string, defaultVal int) int {
	valueStr := getEnv(name, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}

	return defaultVal
}

// Helper to read an environment variable into a bool or return default value
func getEnvAsBool(name string, defaultVal bool) bool {
	valStr := getEnv(name, "")
	if val, err := strconv.ParseBool(valStr); err == nil {
		return val
	}

	return defaultVal
}

// Helper to read an environment variable into a string slice or return default value
func getEnvAsSlice(name string, defaultVal []string, sep string) []string {
	valStr := getEnv(name, "")

	if valStr == "" {
		return defaultVal
	}

	val := strings.Split(valStr, sep)

	return val
}
