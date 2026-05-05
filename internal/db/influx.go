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

	// Critical level
	if temp >= 38.0 {
		log.Println("[ALERT] CRITICAL threshold of 38°C breached! Sending email immediately.")
		alert.SendEmailAlert(temp, "critical")
		writeAlertEvent(temp, "critical")
		return
	}

	// Warning level
	if temp >= 32.0 {
		if shouldSendWarningAlert() {
			log.Println("[ALERT] WARNING threshold of 32°C breached! Sending email.")
			alert.SendEmailAlert(temp, "warning")
			writeAlertEvent(temp, "warning")
		} else {
			log.Println("[ALERT] Temperature >= 32°C detected but currently in cooldown (30 minutes). Skipping email.")
		}
	}
}

func shouldSendWarningAlert() bool {
	org := os.Getenv("INFLUX_ORG")
	bucket := os.Getenv("INFLUX_BUCKET")
	queryAPI := client.QueryAPI(org)

	// Flux query to check for warning alerts in the last 30 minutes
	query := fmt.Sprintf(`
		from(bucket:"%s")
			|> range(start: -30m)
			|> filter(fn: (r) => r._measurement == "alert")
			|> filter(fn: (r) => r.level == "warning")
			|> count()
	`, bucket)

	result, err := queryAPI.Query(context.Background(), query)
	if err != nil {
		log.Printf("[INFLUX] Error querying alerts: %v\n", err)
		return true // Fallback to send alert if query fails
	}

	for result.Next() {
		val := result.Record().Value()
		if count, ok := val.(int64); ok && count > 0 {
			return false
		}
	}

	if result.Err() != nil {
		log.Printf("[INFLUX] Query parsing error: %v\n", result.Err())
	}

	return true
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
