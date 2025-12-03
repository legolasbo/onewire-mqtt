package main

import (
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type Daemon struct {
	Config  Config
	Logger  Logger
	client  mqtt.Client
	sensors []Sensor
}

func (d *Daemon) Start() {
	d.Logger.Info("Starting daemon...")

	var err error
	d.client, err = getMQTTClient(d.Config, d.Logger)
	if err != nil {
		log.Fatalln(err)
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	updateTicker := time.NewTicker(time.Duration(d.Config.UpdatePeriod) * time.Second)

	for {
		select {
		case <-signals:
			return
		case <-updateTicker.C:
			go d.readSensors()
			break
		}
	}
}

func (d *Daemon) readSensors() {
	for _, sensor := range d.sensors {
		value, err := os.ReadFile("/sys/bus/w1/devices/" + sensor.SensorId + "/temperature")
		if err != nil {
			d.Logger.Error(err.Error())
			continue
		}

		stringValue := string(value[:])
		stringValue = strings.Trim(stringValue, "\n")

		c, err := strconv.ParseFloat(stringValue, 64)
		if err != nil {
			d.Logger.Error(err.Error())
			continue
		}

		d.publish(fmt.Sprintf("home/sensors/temperature/%s", sensor.Alias), fmt.Sprintf("%.2f", c/1000))
	}
}

func (d *Daemon) publish(topic string, msg string) mqtt.Token {
	d.Logger.Debug(fmt.Sprintf("Publishing to %s: %s", topic, msg))
	return d.client.Publish(topic, 1, false, msg)
}

func (d *Daemon) publishAndWait(topic string, msg string) {
	token := d.publish(topic, msg)
	token.Wait()
}

func createDaemon(config Config, logger Logger) Daemon {
	d := Daemon{Config: config, Logger: logger}

	ids, err := LocateConnectedSensors(logger)
	if err != nil {
		log.Fatalln(err)
	}

	d.sensors = validateSensors(loadSensorConfiguration(config, ids))
	logger.Info("Sensors loaded")

	return d
}

func validateSensors(sensors []Sensor) []Sensor {
	for _, sensor := range sensors {
		if sensor.Alias == "" {
			sensor.Alias = sensor.SensorId
		}
		if sensor.MqttId == "" {
			sensor.MqttId = sensor.SensorId
		}
	}
	return sensors
}

func loadSensorConfiguration(config Config, ids []string) []Sensor {
	sensorMap := make(map[string]Sensor, len(ids))
	for _, id := range ids {
		sensorMap[id] = Sensor{SensorId: id, MqttId: id, Alias: id}
	}

	for _, sensor := range config.Sensors {
		_, knownSensor := sensorMap[sensor.SensorId]
		if knownSensor {
			sensorMap[sensor.SensorId] = sensor
		}
	}

	sensors := make([]Sensor, 0, len(sensorMap))
	for _, sensor := range sensorMap {
		sensors = append(sensors, sensor)
	}

	return sensors
}
