package sensor

import (
	"encoding/json"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/versenilvis/templog-monitoring/internal/hub"
)

type MQTTPayload struct {
	Temp float64 `json:"temp"`
	Hum  float64 `json:"hum"`
}

func ReadMQTT(h *hub.Hub) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker("tcp://broker.emqx.io:1883")
	opts.SetClientID("templog_backend")
	opts.SetAutoReconnect(true)
	opts.SetConnectRetry(true)
	opts.SetConnectRetryInterval(time.Second * 5)

	opts.OnConnectionLost = func(c mqtt.Client, err error) {
		log.Printf("[MQTT] Connection lost: %v", err)
	}

	opts.OnConnect = func(c mqtt.Client) {
		log.Println("[MQTT] Connected to EMQX Cloud Broker")
		log.Println("[MQTT] Subscribing to topic: room/sensor/data")

		if token := c.Subscribe("room/sensor/data", 1, nil); token.Wait() && token.Error() != nil {
			log.Printf("[MQTT] Subscribe error: %v", token.Error())
		}
	}

	client := mqtt.NewClient(opts)

	client.AddRoute("room/sensor/data", func(c mqtt.Client, m mqtt.Message) {
		var payload MQTTPayload
		if err := json.Unmarshal(m.Payload(), &payload); err != nil {
			log.Printf("[MQTT] Payload parse error: %v | Raw: %s", err, string(m.Payload()))
			return
		}

		log.Printf("[MQTT] Received: Temp=%.2f, Hum=%.2f", payload.Temp, payload.Hum)

		h.Broadcast(hub.SensorData{
			Temperature: payload.Temp,
			Humidity:    payload.Hum,
			Timestamp:   time.Now(),
		})
	})

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Printf("[MQTT] Initial connect error: %v", token.Error())
	}

	select {}
}
