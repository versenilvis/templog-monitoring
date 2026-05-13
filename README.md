<div align="center">
  <img width=10% alt="ce" src="https://github.com/user-attachments/assets/b98ff0ac-1f45-4ff7-aa04-ab652528cbdd" />
  <h1>Templog-monitoring</h1>
  <p><b>Hệ thống giám sát nhiệt độ và độ ẩm phòng thời gian thực qua ESP32, MQTT, backend Go và frontend WebSocket</b></p>
  <p><a href="#cài-đặt"><strong>Cài đặt »</strong></a></p>
  <img src="https://camo.githubusercontent.com/b16ecdcac9c3d21ec3a49459430f747b46b3a37acc95ee468d87d0ec61ff2392/68747470733a2f2f692e696d6775722e636f6d2f576d4d6e5352742e706e67">
</div>

<table>
  <td><img src="https://github.com/user-attachments/assets/6152bb4c-d080-41de-b429-b473144db4cb"></td>
  <td><img src="https://github.com/user-attachments/assets/13049b8e-dad2-4678-9451-a29d186a01df"></td>
</table>

> [!IMPORTANT]
> Nhánh này sử dụng WiFi và MQTT. ESP32 đã có WiFi tích hợp sẵn, không cần thêm phần cứng nào khác.  
> Chuyển sang nhánh [`firmware-usb`](https://github.com/versenilvis/templog-monitoring/tree/firmware-usb) nếu bạn muốn dùng phiên bản USB/UART

## Cách thức hoạt động

ESP32 khởi động và kết nối WiFi. Nếu không tìm thấy thông tin xác thực, thiết bị sẽ vào chế độ **SmartConfig** (dùng app ESPTouch v2 để cấu hình). Sau khi kết nối thành công, nó sẽ kết nối đến **EMQX Cloud Broker** và publish `{"temp": xx.xx, "hum": xx.xx}` lên topic mỗi 2 giây.

Server Go kết nối đến cùng cloud broker, phân tích payload JSON và phát dữ liệu qua WebSocket để hiển thị lên web.

Vậy điểm nổi bật của sản phẩm này là gì? Vì hệ thống và firmware kết nối qua cloud, vị trí địa lý không còn là vấn đề. Chỉ cần cấu hình đúng, server và firmware của bạn có thể ở bất kỳ đâu trên thế giới, miễn là có WiFi để kết nối đến cloud.

## Cần chuẩn bị

- [Bun](https://bun.com/)
- [Golang](https://go.dev/)
- [Docker](https://www.docker.com/) (vui lòng đọc thêm về [Docker Compose](https://docs.docker.com/compose/))
- [Make](https://www.gnu.org/software/make/) (dùng cho Makefile)
- [EIM](https://docs.espressif.com/projects/idf-im-ui/en/latest/)
- [ESP-IDF](https://docs.espressif.com/projects/esp-idf/en/v3.1.5/get-started/linux-setup.html)

> [!NOTE]
> Sau khi cài ESP-IDF, bạn nên tạo alias trong file cấu hình shell của mình.  
> Vì nếu phải source mỗi lần mở terminal mới, sẽ phải mất một khoản thời gian ngắn để load config.

Ví dụ:
```bash
alias idf="source /opt/esp-idf/export.sh"
```

## Cài đặt

- Nếu đã có Make, chạy lệnh sau để thực thi [`./scripts/setup.sh`](./scripts/setup.sh):
```bash
make setup
```
Hoặc chạy lệnh này nếu chưa cài Make:
```bash
chmod +x scripts/setup.sh && bash scripts/setup.sh
```

> [!WARNING]
> **Script này chỉ hỗ trợ Linux, macOS hoặc Windows WSL2. Nếu muốn chạy trên Windows thuần, vui lòng đọc script và chuyển đổi thủ công sang cú pháp PowerShell.**  
> Hiện chỉ hỗ trợ `pacman`, `apt`, `dnf`.  
> Nếu dùng distro khác, hãy thay lệnh bằng package manager tương ứng.  

## Cấu hình

Xem file `.env.example`, tạo file `.env` với thông tin xác thực của bạn.

1. Với email, nên dùng Gmail App Passwords:
   - Truy cập: `https://myaccount.google.com/apppasswords`
   - Đặt tên cho ứng dụng
   - Sao chép mật khẩu được tạo và dán vào `SMTP_PASSWORD` trong file `.env`

2. Với database, `INFLUX_TOKEN` và `DOCKER_INFLUXDB_INIT_ADMIN_TOKEN` là như nhau.
   - Chạy lệnh sau để tạo token rồi dán vào cả hai:
```bash
openssl rand -hex 32
```

## Cách chạy

- Đầu tiên, build firmware:
```bash
make build
```
- Tiếp theo, flash firmware (chỉ cần làm một lần):
```bash
make flash
```
- Sau đó, monitor qua ESP-IDF (chỉ cần làm một lần):
```bash
make monitor
```

## Giám sát qua web

- Chạy container Docker InfluxDB và khởi động cả web lẫn server bằng một lệnh duy nhất:
```bash
make app
```

*Đừng quên dùng `make down` để tắt DB khi dừng tất cả dịch vụ, vì khi kill web/server, container Docker InfluxDB vẫn đang chạy nền.*

## Showcase khắc mạch in PCB (PCB Etching)

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

## Lời cảm ơn

Cảm ơn tất cả các thành viên đã đóng góp cho dự án này.

- 24520598 - Mai Thế Hùng
- 24520560 - Nguyễn Trọng Hoàng
- 24520537 - Lã Minh Hoàng
- 24520563 - Phạm Việt Hoàng
- 24520570 - Trần Nguyễn Huy Hoàng
- 24521016 - Trần Nguyễn Hoàng Long

## Giấy phép

Dự án này được cấp phép theo [Giấy phép 0BSD](LICENSE). Nghĩa là bạn có thể làm bất cứ điều gì với nó.
