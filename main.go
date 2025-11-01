package main

import (
	"log"
	"os"
	"strings"
)

const DefaultConfigPath = "/etc/onewire-mqtt/config.yml"

func main() {
	configPath := DefaultConfigPath
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
	log.Printf("[INFO]: Read slaves: %s", string(data))

	sensors := strings.Split(string(data), "\n")
	log.Printf("[INFO]: Found sensors: %s", sensors)
	if len(sensors) > 0 {
		sensors = sensors[:len(sensors)-1]
	}

	return sensors, nil
}
