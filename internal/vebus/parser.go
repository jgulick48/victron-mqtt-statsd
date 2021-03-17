package vebus

import (
	"encoding/json"
	"fmt"
	"github.com/jgulick48/victron-mqtt-statsd/internal/metrics"
	"log"
	"strings"
)

func sendGaugeMetric(name string, tags []string, value float64) {
	if metrics.StatsEnabled {
		err := metrics.Metrics.Gauge(name, value, tags, 1)
		if err != nil {
			log.Printf("Got error trying to send metric %s", err.Error())
		}
	}
	log.Printf("got metric %s with tags %s and value %v", name, tags, value)
}

func formatTag(key, value string) string {
	return fmt.Sprintf("%s:%s", key, value)
}

func ProcessData(topic string, message []byte) error {
	var payload Message
	err := json.Unmarshal(message, &payload)
	if err != nil {
		return err
	}
	segments := strings.Split(topic, "/")
	parser := GetDataParser(segments)
	parser(segments, payload)
	return nil
}

func ParseACData(segments []string, message Message) {
	if !message.Value.Valid {
		return
	}
	tags := []string{
		formatTag("deployment", segments[1]),
		formatTag("vebus.id", segments[3]),
	}
	var metricName string
	var shouldSend bool
	switch segments[5] {
	case "ActiveIn", "Out":
		tags, metricName, shouldSend = parseACLineMeasurements(tags, segments)
	}
	if metricName != "" && shouldSend {
		sendGaugeMetric(metricName, tags, message.Value.Float64)
	}
}

func parseACLineMeasurements(tags []string, segments []string) ([]string, string, bool) {
	if len(segments) == 8 {
		tags = append(tags, formatTag("line", segments[6]))
	}
	unit := ""
	switch segments[len(segments)-1] {
	case "F":
		unit = "frequency"
	case "I":
		unit = "current"
	case "P":
		unit = "power"
	case "V":
		unit = "volts"
	}
	if unit == "" {
		return tags, "", false
	}
	return tags, fmt.Sprintf("ac_%s_%s", strings.ToLower(segments[5]), unit), true
}

func DefaultParser(segments []string, message Message) {
	return
}

func GetDataParser(segments []string) func(topic []string, message Message) {
	if len(segments) < 5 {
		return DefaultParser
	}
	switch segments[4] {
	case "Ac":
		return ParseACData
	default:
		return DefaultParser
	}
}
