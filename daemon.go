package main

import (
	"encoding/json"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"log"
	"os"
	"os/signal"
	"strconv"
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

	expireTicker := time.NewTicker(time.Duration(d.Config.UpdatePeriod*100) * time.Second)
	updateTicker := time.NewTicker(time.Duration(d.Config.UpdatePeriod) * time.Second)

	d.configureHomeAssistant()
	for {
		select {
		case <-signals:
			return
		case <-expireTicker.C:
			go d.configureHomeAssistant()
			break
		case <-updateTicker.C:
			go d.readSensors()
			break
		}
	}
}

func (d *Daemon) configureHomeAssistant() {
	if !d.Config.HomeAssistant {
		return
	}

	for _, sensor := range d.sensors {
		d.publishHomeAssistantConfig(sensor.HomeAssistantConfig(d.Config))
	}
}

func (d *Daemon) publishHomeAssistantConfig(topic string, config HomeAssistantConfig) {
	val, err := json.Marshal(config)
	if err != nil {
		d.Logger.Warn(err.Error())
		return
	}
	d.publish(topic, string(val))
}

func (d *Daemon) readSensors() {
	for _, sensor := range d.sensors {
		value, err := os.ReadFile("/sys/bus/w1/devices/" + sensor.SensorId + "/temperature")
		if err != nil {
			d.Logger.Error(err.Error())
			continue
		}

		c, err := strconv.ParseFloat(string(value[:]), 64)
		if err != nil {
			d.Logger.Error(err.Error())
			continue
		}

		d.publish(fmt.Sprintf("%s/%s/%s/%s", d.Config.Prefix, d.Config.ClientId, "temperature", sensor), fmt.Sprintf("%.2f", c))
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

	d.sensors = loadSensorConfiguration(config, ids)
	validateSensors(d.sensors)

	return d
}

func validateSensors(sensors []Sensor) {
	for _, sensor := range sensors {
		if sensor.Alias == "" {
			sensor.Alias = sensor.SensorId
		}
		if sensor.MqttId == "" {
			sensor.MqttId = sensor.SensorId
		}
	}
}

func loadSensorConfiguration(config Config, ids []string) []Sensor {
	var sensorMap map[string]Sensor
	for _, id := range ids {
		sensorMap[id] = Sensor{SensorId: id}
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
