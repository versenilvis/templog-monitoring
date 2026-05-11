<div align="center">
  <img width=10% alt="ce" src="https://github.com/user-attachments/assets/b98ff0ac-1f45-4ff7-aa04-ab652528cbdd" />
  <h1>Templog-monitoring</h1>
  <p><b>A real-time temperature and humidity monitor ESP32 over MQTT, Go backend, WebSocket frontend</b></p>
  <p><a href="#installation"><strong>Installation »</strong></a></p>
  <img src="https://camo.githubusercontent.com/b16ecdcac9c3d21ec3a49459430f747b46b3a37acc95ee468d87d0ec61ff2392/68747470733a2f2f692e696d6775722e636f6d2f576d4d6e5352742e706e67">
</div>

<table>
  <td ><img src="https://github.com/user-attachments/assets/6152bb4c-d080-41de-b429-b473144db4cb"></td>
  <td ><img src="https://github.com/user-attachments/assets/13049b8e-dad2-4678-9451-a29d186a01df"></td>
</table>

>[!IMPORTANT]
> This branch uses WiFi and MQTT. ESP32 has built-in WiFi, no extra hardware needed  
> Switch to the [`firmware-usb`](https://github.com/versenilvis/templog-monitoring/tree/firmware-usb) branch if you prefer the USB/UART version

## Under the hood
The ESP32 boots and connects to WiFi. If no credentials are found, it enters **SmartConfig** mode (use ESPTouch v2 app to provision). Once connected, it connects to the **EMQX Cloud Broker** and publishes `{"temp": xx.xx, "hum": xx.xx}` to the topic every 2 seconds.

The Go server connects to the same cloud broker, parses the JSON payload, and broadcasts it via WebSocket.

## Dependencies
- [Bun](https://bun.com/)
- [Golang](https://go.dev/)
- [Docker](https://www.docker.com/) (please also read [Docker Compose](https://docs.docker.com/compose/))
- [Make](https://www.gnu.org/software/make/) (for Makefile)
- [EIM](https://docs.espressif.com/projects/idf-im-ui/en/latest/)
- [ESP-IDF](https://docs.espressif.com/projects/esp-idf/en/v3.1.5/get-started/linux-setup.html)

> [!NOTE]
> After installing ESP-IDF, you should make an alias for it in your shell config file  
> Because if you source it everytime you open a new terminal, it will take a short time to load it

E.g.
```bash
alias idf="source /opt/esp-idf/export.sh"
```

## Installation
- Simply run [`./scripts/setup.sh`](./scripts/setup.sh) with this command if you already have Make:
```bash
make setup
```
Or run this if you've not installed it yet:
```bash
chmod +x scripts/setup.sh && bash scripts/setup.sh
```
> [!WARNING]
> **This script is Linux, macOS or Windows WSL2 only. If you want to run on native Windows, please read the script and convert to Windows powershell syntax to install and configure manually**  
> Currently only support `pacman`, `apt`, `dnf`  
> If you use another distro, remember to change the command to your distro's package manager
<!--
- Install all necessary packages
```bash
make setup
```

- Install Avahi

`sudo pacman -S mosquitto avahi`

- Configure Mosquitto (MQTT Broker)


Open file: `sudo nano /etc/mosquitto/mosquitto.conf` (or use your own editor)

```
listener 1883
allow_anonymous true
```

- Configure Avahi (mDNS Service Discovery)

Open file: `sudo nano /etc/avahi/services/mqtt.service`

```xml
<?xml version="1.0" standalone='no'?>
<!DOCTYPE service-group SYSTEM "avahi-service.dtd">
<service-group>
  <name>MQTT Broker</name>
  <service>
    <type>_mqtt._tcp</type>
    <port>1883</port>
  </service>
</service-group>
```

- Start services
```bash
sudo systemctl start mosquitto
sudo systemctl start avahi-daemon
```

- Check

```bash
# Check broker is running
sudo systemctl status mosquitto

# Check avahi is advertising
avahi-browse _mqtt._tcp
```
Expected:
```
+ wlp1s0 IPv4 MQTT Broker _mqtt._tcp local
```
-->
## Configuration
Look at the `.env.example` file, create `.env` file with your own credentials
1. With email, I recommend you to use gmail app passwords
   - Go to: `https://myaccount.google.com/apppasswords`
   - Name your application
   - Copy generated password and paste it to `SMTP_PASSWORD` in `.env` file
2. With the database, `INFLUX_TOKEN` and `DOCKER_INFLUXDB_INIT_ADMIN_TOKEN` are the same
   - Therefore, simply run this command to generate token and paste in both: 
```bash
openssl rand -hex 32
```

## How to run
- First, build the firmware
```bash
make build
```
- Then, flash the firmware (you only need to do this once)
```bash
make flash
```
- Then you need to monitor it via ESP-IDF (you only need to do this once)
```bash
make monitor
```
## Monitoring via the web
- Run the InfluxDB database docker container and run both web and server in one single command
```bash
make app
```

*Don't forget to use `make down` to turn off DB when you turn off all services, because when we kill web/server, the InfluxDB docker container is still running in the background.*

<div align="center">
  
</div>

## PCB Etching showcase

<div align="center">
  <table>
    <tr>
      <td align="center"><img src="https://github.com/user-attachments/assets/ec5406e6-78d8-4c76-b587-d04c93455509" width="250"/></td>
      <td align="center"><img src="https://github.com/user-attachments/assets/7a390c2c-73d5-4a92-8c15-f42b83440af6" width="250"/></td>
      <td align="center"><img src="https://github.com/user-attachments/assets/19734409-0c37-4f56-961e-d80f7ac1615d" width="250"/></td>
    </tr>
  </table>

  <table>
    <tr>
      <td align="center"><img src="https://github.com/user-attachments/assets/db1ec918-5fc8-4e9f-81db-563997aa6dea" width="250"/></td>
      <td align="center"><img src="https://github.com/user-attachments/assets/1ce565ee-52be-4e09-9335-24c67c76543d" width="250"/></td>
    </tr>
  </table>
</div>
<!--
## How to quickly view your data in InfluxDB
- Open `localhost:8086`
- Sign in with your username and password from `.env`
- Open Data Explorer tab
- Choose all records you need to view
<img width="2350" height="757" alt="image" src="https://github.com/user-attachments/assets/79f96129-a6f6-4cf9-bd08-bbb4fb753435" />
- Choose table
<img width="374" height="523" alt="image" src="https://github.com/user-attachments/assets/ea01b320-25e9-4641-9e8c-03f9aa638a82" />
- Click submit
<img width="2382" height="1475" alt="image" src="https://github.com/user-attachments/assets/77158693-bc2b-48dd-ba97-9e242e70bd2e" />
-->
