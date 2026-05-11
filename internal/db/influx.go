package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/versenilvis/templog-monitoring/internal/alert"
)

var client influxdb2.Client

func Init() {
	url := os.Getenv("INFLUX_URL")
	token := os.Getenv("INFLUX_TOKEN")
	if url == "" || token == "" {
		log.Println("[INFLUX] Missing INFLUX_URL or INFLUX_TOKEN in .env")
		return
	}
	client = influxdb2.NewClient(url, token)
	log.Println("[INFLUX] Connected to InfluxDB")
}

func WriteSensorData(temp float64, hum float64) {
	if client == nil {
		return
	}
	org := os.Getenv("INFLUX_ORG")
	bucket := os.Getenv("INFLUX_BUCKET")
	writeAPI := client.WriteAPIBlocking(org, bucket)

	p := influxdb2.NewPointWithMeasurement("sensor").
		AddTag("room", "classroom").
		AddField("temperature", temp).
		AddField("humidity", hum).
		SetTime(time.Now())

	if err := writeAPI.WritePoint(context.Background(), p); err != nil {
		log.Printf("[INFLUX] Error writing data: %v\n", err)
	}
}

func ProcessAlert(temp float64) {
	if client == nil {
		return
	}

	// 1. Critical level (38°C) - 2min Cooldown
	if temp >= 38.0 {
		shouldSend, _ := checkAlertCooldown("critical", 2)
		if shouldSend {
			log.Println("[ALERT] CRITICAL threshold of 38°C breached! Sending email.")
			alert.SendEmailAlert(temp, "critical")
			writeAlertEvent(temp, "critical")
		}
		return
	}

	// 2. Reset condition (Temp < 30) - Clears the need for cooldown
	if temp < 30.0 {
		if isLastEventAlert() {
			log.Println("[ALERT] Temperature cooled down below 30°C. Alert system reset and re-armed.")
			writeAlertEvent(temp, "reset")
		}
		return
	}

	// 3. Warning level (35°C) - Hybrid: Reset OR 30min Cooldown
	if temp >= 35.0 {
		shouldSend, reason := checkAlertCooldown("warning", 30)
		if shouldSend {
			log.Printf("[ALERT] WARNING threshold of 35°C breached! (%s). Sending email.\n", reason)
			alert.SendEmailAlert(temp, "warning")
			writeAlertEvent(temp, "warning")
		}
	}
}

func checkAlertCooldown(level string, cooldownMinutes int) (bool, string) {
	org := os.Getenv("INFLUX_ORG")
	bucket := os.Getenv("INFLUX_BUCKET")
	queryAPI := client.QueryAPI(org)

	// Get the last event across all alert types to check the sequence
	query := fmt.Sprintf(`
		from(bucket:"%s")
			|> range(start: -24h)
			|> filter(fn: (r) => r._measurement == "alert")
			|> group()
			|> last()
	`, bucket)

	result, err := queryAPI.Query(context.Background(), query)
	if err != nil {
		return true, "query error"
	}

	if result.Next() {
		lastLevel := result.Record().ValueByKey("level").(string)
		lastTime := result.Record().Time()

		// If the system was reset, we always send the next warning/critical
		if lastLevel == "reset" {
			return true, "system was reset"
		}

		// If we are checking for a different (higher) level than the last alert, send immediately
		if level == "critical" && lastLevel == "warning" {
			return true, "escalated to critical"
		}

		// If it's the same level or lower, check the cooldown
		if time.Since(lastTime) > time.Duration(cooldownMinutes)*time.Minute {
			return true, "cooldown expired"
		}

		return false, ""
	}

	return true, "no previous data"
}

func isLastEventAlert() bool {
	org := os.Getenv("INFLUX_ORG")
	bucket := os.Getenv("INFLUX_BUCKET")
	queryAPI := client.QueryAPI(org)

	query := fmt.Sprintf(`
		from(bucket:"%s")
			|> range(start: -24h)
			|> filter(fn: (r) => r._measurement == "alert")
			|> group()
			|> last()
	`, bucket)

	result, err := queryAPI.Query(context.Background(), query)
	if err != nil {
		return false
	}

	if result.Next() {
		level := result.Record().ValueByKey("level").(string)
		return level == "warning" || level == "critical"
	}
	return false
}

func writeAlertEvent(temp float64, level string) {
	org := os.Getenv("INFLUX_ORG")
	bucket := os.Getenv("INFLUX_BUCKET")
	writeAPI := client.WriteAPIBlocking(org, bucket)

	p := influxdb2.NewPointWithMeasurement("alert").
		AddTag("level", level).
		AddField("temp", temp).
		SetTime(time.Now())

	if err := writeAPI.WritePoint(context.Background(), p); err != nil {
		log.Printf("[INFLUX] Error writing alert event: %v\n", err)
	}
}

type DailyStats struct {
	MinTemp float64 `json:"min_temp"`
	MaxTemp float64 `json:"max_temp"`
	MinHum  float64 `json:"min_hum"`
	MaxHum  float64 `json:"max_hum"`
}

func GetTodayStats() (DailyStats, error) {
	// Initialize with extreme values to ensure correct min/max calculation
	stats := DailyStats{
		MinTemp: 99.0,
		MaxTemp: -99.0,
		MinHum:  100.0,
		MaxHum:  0.0,
	}

	if client == nil {
		return stats, fmt.Errorf("influxdb client not initialized")
	}

	org := os.Getenv("INFLUX_ORG")
	bucket := os.Getenv("INFLUX_BUCKET")
	queryAPI := client.QueryAPI(org)

	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	startStr := startOfDay.Format(time.RFC3339)

	query := fmt.Sprintf(`
		data = from(bucket:"%s")
			|> range(start: %s)
			|> filter(fn: (r) => r._measurement == "sensor")
			|> filter(fn: (r) => r._field == "temperature" or r._field == "humidity")

		data |> group(columns: ["_field"]) |> min() |> yield(name: "min")
		data |> group(columns: ["_field"]) |> max() |> yield(name: "max")
	`, bucket, startStr)

	result, err := queryAPI.Query(context.Background(), query)
	if err != nil {
		return stats, err
	}

	hasData := false
	for result.Next() {
		hasData = true
		field := result.Record().Field()
		value, ok := result.Record().Value().(float64)
		if !ok {
			continue
		}

		// Use result name from the record to distinguish min/max yields
		op := result.Record().Result()

		switch op {
		case "min":
			switch field {
			case "temperature":
				stats.MinTemp = value
			case "humidity":
				stats.MinHum = value
			}
		case "max":
			switch field {
			case "temperature":
				stats.MaxTemp = value
			case "humidity":
				stats.MaxHum = value
			}
		}
	}

	// If no data found, reset to 0
	if !hasData {
		return DailyStats{}, nil
	}

	return stats, result.Err()
}
