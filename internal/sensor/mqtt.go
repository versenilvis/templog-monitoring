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
	// 1. SỬA CHỖ NÀY: Trỏ lên Cloud EMQX thay vì localhost
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
		log.Println("[MQTT] Subscribing to topic: uit/ce103/project/nhietdovadoamphong")
		
		// 2. SỬA CHỖ NÀY: Lắng nghe đúng cái Topic đồ án
		if token := c.Subscribe("uit/ce103/project/nhietdovadoamphong", 1, nil); token.Wait() && token.Error() != nil {
			log.Printf("[MQTT] Subscribe error: %v", token.Error())
		}
	}

	client := mqtt.NewClient(opts)

	// 3. SỬA CHỖ NÀY: Định tuyến (Route) để hứng data từ Topic đồ án
	client.AddRoute("uit/ce103/project/nhietdovadoamphong", func(c mqtt.Client, m mqtt.Message) {
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
