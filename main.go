package main

import (
	"github.com/DataDog/datadog-go/statsd"
	"github.com/jgulick48/victron-mqtt-statsd/internal/metrics"
	"github.com/jgulick48/victron-mqtt-statsd/internal/mqtt"
	"log"
	"os"
)

func main() {
	var err error
	metrics.Metrics, err = statsd.New("127.0.0.1:8125")
	if err == nil {
		metrics.StatsEnabled = true
	}
	log.Printf("Connecing to host %s with device %s", os.Getenv("MQTT_HOST"), os.Getenv("DEVICE_ID"))
	config := mqtt.Configuration{
		Host:     os.Getenv("MQTT_HOST"),
		Port:     1883,
		DeviceID: os.Getenv("DEVICE_ID"),
	}
	client := mqtt.NewClient(config)
	client.Connect()
}
