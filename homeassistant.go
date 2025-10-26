package main

import (
	"fmt"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const DefaultExpireAfter = 60

type HomeAssistantConfig struct {
	Name              string              `json:"name,omitempty"`
	DeviceClass       string              `json:"device_class,omitempty"`
	UnitOfMeasurement string              `json:"unit_of_measurement,omitempty"`
	Device            HomeAssistantDevice `json:"device"`
	ExpireAfter       int                 `json:"expire_after,omitempty"`
	StateTopic        string              `json:"state_topic,omitempty"`
	UniqueId          string              `json:"unique_id,omitempty"`
	ObjectId          string              `json:"object_id,omitempty"`
	StateClass        string              `json:"state_class,omitempty"`
	Icon              string              `json:"icon,omitempty"`
}

type HomeAssistantDevice struct {
	Name        string `json:"name,omitempty"`
	Model       string `json:"model,omitempty"`
	Identifiers string `json:"identifiers,omitempty"`
}

func GetHomeAssistantDevice(conf Config) HomeAssistantDevice {
	caser := cases.Title(language.Dutch)
	return HomeAssistantDevice{
		Name:        caser.String(conf.ClientId),
		Model:       conf.ClientId,
		Identifiers: conf.ClientId,
	}
}

func (s Sensor) HomeAssistantConfig(config Config) (string, HomeAssistantConfig) {
	uniqueId := fmt.Sprintf("%s_%s", config.ClientId, s.SensorId)
	topic := fmt.Sprintf("homeassistant/sensor/%s/%s/config", config.ClientId, s.SensorId)

	return topic, HomeAssistantConfig{
		Name:              s.Alias,
		DeviceClass:       "temperature",
		UnitOfMeasurement: "°C",
		Device:            GetHomeAssistantDevice(config),
		ExpireAfter:       DefaultExpireAfter,
		StateTopic:        fmt.Sprintf("%s/%s/%s/%s", config.Prefix, config.ClientId, "temperature", s.SensorId),
		UniqueId:          uniqueId,
		ObjectId:          uniqueId,
		StateClass:        "MEASUREMENT",
		Icon:              "mdi:temperature-celsius",
	}
}
