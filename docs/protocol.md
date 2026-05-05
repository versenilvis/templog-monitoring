# Communication Protocol

## 1. ESP32 to Broker (MQTT)
- **Broker**: `broker.emqx.io:1883` (Cloud)
- **Topic**: `room/sensor/data`
- **Payload**: `{"temp": 26.45, "hum": 58.20}` (JSON)

## 2. Server to Client (WebSocket)
- **Endpoint**: `ws://localhost:8080/ws`
- **Format**:
```json
{
  "type": "history | live",
  "data": "SensorData | []SensorData"
}
```

- **history**: Array of last 60 points (sent on connect).
- **live**: Single data point (sent every 2s).

## 3. WiFi Setup
- **SmartConfig**: Use **ESPTouch v2** app to provision WiFi credentials. No mDNS or hardcoded SSID required.
