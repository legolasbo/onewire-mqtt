package main

import (
	"os"
	"strings"
)

const DEFAULT_CONFIG_PATH = "/etc/onewire-mqtt/config.yml"

func main() {
	configPath := DEFAULT_CONFIG_PATH
	if len(os.Args) > 1 {
		configPath = os.Args[len(os.Args)-1]
	}

	config := loadConfig(configPath)
	logger := createLogger(config)

	daemon := createDaemon(config, logger)

	daemon.Start()
}

// LocateConnectedSensors get all connected sensor IDs as array
func LocateConnectedSensors(logger Logger) ([]string, error) {
	data, err := os.ReadFile("/sys/bus/w1/devices/w1_bus_master1/w1_master_slaves")
	if err != nil {
		logger.Error("Unable to detect onewire sensor id's, are you sure you've enabled the w1-gpio overlay?")
		return nil, err
	}

	sensors := strings.Split(string(data), "\n")
	if len(sensors) > 0 {
		sensors = sensors[:len(sensors)-1]
	}

	return sensors, nil
}
