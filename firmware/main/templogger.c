#include "esp_event.h" // Thư viện quản lý sự kiện (như khi kết nối/mất WiFi)
#include "esp_log.h"   // Thư viện in log màu sắc ra Terminal (thay cho printf)
#include "esp_netif.h" // Lớp giao diện mạng (chuyển đổi dữ liệu thô sang gói tin mạng)
#include "esp_wifi.h"          // Thư viện điều khiển driver WiFi của ESP32
#include "freertos/FreeRTOS.h" // Hệ điều hành thời gian thực (RTOS)
#include "freertos/task.h"     // Quản lý đa nhiệm (Task)
#include "mqtt_client.h"       // Giao thức truyền tin MQTT (Publish/Subscribe)
#include "nvs_flash.h" // Bộ nhớ không tự xóa (dùng để lưu cấu hình WiFi)

// BỔ SUNG: Thư viện SmartConfig để hứng mật khẩu từ App ESPTouch v2 (Thay cho
// Provisioning)
#include "esp_smartconfig.h"

#include <stddef.h>
#include <stdint.h>
#include <stdio.h>
#include <string.h>

// Header linh kiện
#include "driver/i2c.h" // Driver điều khiển ngoại vi I2C
#include "oledi2c.h"    // Thư viện màn hình OLED
#include "sht30.h"      // Thư viện cảm biến nhiệt độ SHT30

static const char *TAG = "TEMP_LOG"; // Tên thẻ để lọc log trong Terminal
static int s_retry_num = 0;          // Biến đếm số lần thử kết nối lại WiFi

// 1. CẤU HÌNH THÔNG SỐ
#define I2C_MASTER_SCL_IO 22      // Chân SCL nối vào GPIO 22
#define I2C_MASTER_SDA_IO 21      // Chân SDA nối vào GPIO 21
#define I2C_MASTER_NUM I2C_NUM_0  // Sử dụng bộ I2C số 0 của ESP32
#define I2C_MASTER_FREQ_HZ 100000 // Tốc độ I2C (100kHz)

// 2. BITMAP
// 2.1 Icon ống nhiệt kế 32x32 pixel = 128 byte
const uint8_t icon_temp_32x32[128] = {
    0x00, 0x00, 0x60, 0x60, 0x60, 0x60, 0x00, 0x00, 0xfc, 0x06, 0x03, 0x01,
    0x03, 0xfe, 0xfc, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
    0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x60, 0x60,
    0x69, 0x69, 0x09, 0x00, 0xff, 0x00, 0x00, 0xe0, 0x00, 0xff, 0xff, 0x00,
    0x00, 0xf8, 0xfc, 0x06, 0x06, 0x06, 0x06, 0x0c, 0x1c, 0x00, 0x00, 0x1c,
    0x16, 0x1e, 0x0c, 0x00, 0x00, 0x00, 0x60, 0x60, 0x69, 0x69, 0x09, 0x00,
    0xff, 0x00, 0x00, 0xff, 0x00, 0xff, 0xff, 0x00, 0x00, 0x07, 0x1f, 0x18,
    0x10, 0x10, 0x10, 0x18, 0x0e, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
    0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x3e, 0x63, 0xcc, 0xde, 0x97,
    0xdc, 0xc1, 0x77, 0x3e, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
    0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00};

// 2.2 Icon giọt nước cho độ ẩm 32x32 pixel = 128 byte
const uint8_t icon_hum_32x32[128] = {
    0x00, 0x00, 0x00, 0x00, 0x00, 0x80, 0xc0, 0xe0, 0x70, 0x3c, 0x1e, 0x0e,
    0x0e, 0x1c, 0x38, 0xf0, 0xe0, 0xc0, 0x80, 0xc0, 0xe0, 0x70, 0x3c, 0x1c,
    0x1c, 0x38, 0xf0, 0xe0, 0xc0, 0x80, 0x00, 0x00, 0x80, 0xe0, 0xf8, 0x7c,
    0x1f, 0x07, 0x03, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
    0x01, 0x03, 0x0f, 0x1f, 0x7d, 0xf0, 0xe0, 0x00, 0x00, 0x00, 0x00, 0x01,
    0x03, 0xdf, 0xfe, 0xf8, 0x3f, 0xff, 0xf3, 0x80, 0x00, 0x00, 0x00, 0x00,
    0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x80, 0xc0, 0xf0, 0x20,
    0x80, 0xff, 0xff, 0x3f, 0x38, 0x38, 0x18, 0x1c, 0x0e, 0x07, 0x03, 0x00,
    0x00, 0x01, 0x03, 0x07, 0x0e, 0x1c, 0x38, 0x38, 0x70, 0x70, 0x70, 0x60,
    0x60, 0x74, 0x76, 0x37, 0x3b, 0x3d, 0x1c, 0x0f, 0x07, 0x03, 0x00, 0x00,
    0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00};

// Handler để điều khiển MQTT
esp_mqtt_client_handle_t mqtt_client = NULL;

// 1. Hàm này nối thẳng lên Cloud, khỏi cần tìm kiếm mDNS lằng nhằng!
static void resolve_server_and_start_mqtt() {
  ESP_LOGI(TAG, "Đã có mạng! Đang kết nối thẳng lên Cloud MQTT...");

  esp_mqtt_client_config_t mqtt_cfg = {
      // Chỉ đích danh địa chỉ máy chủ Cloud
      .broker.address.uri = "mqtt://broker.emqx.io:1883",
      // Đặt tên Client ID cho đỡ trùng với người khác (có thể đổi tên khác tùy
      // ý)
      .credentials.client_id = "esp32_templog_hung_uit",
  };

  if (mqtt_client != NULL) {
    esp_mqtt_client_stop(mqtt_client);
    esp_mqtt_client_destroy(mqtt_client);
  }

  mqtt_client = esp_mqtt_client_init(&mqtt_cfg);
  esp_mqtt_client_start(mqtt_client);
}

/**
 * @brief Handler: Xử lý TẤT CẢ sự kiện từ WiFi đến SmartConfig
 */
static void event_handler(void *arg, esp_event_base_t base, int32_t id,
                          void *data) {
  // 1. Nhóm sự kiện kết nối WiFi thông thường
  if (base == WIFI_EVENT && id == WIFI_EVENT_STA_START) {
    esp_err_t err = esp_wifi_connect();

    // Nếu NVS trống (không có tên WiFi), hàm connect sẽ trả về lỗi
    // ESP_ERR_WIFI_SSID
    if (err == ESP_ERR_WIFI_SSID) {
      ESP_LOGI(TAG,
               "============================================================");
      ESP_LOGI(TAG, ">> [SMARTCONFIG] NÃO TRỐNG! BẬT HỨNG PASS NGAY LẬP TỨC");
      ESP_LOGI(TAG, ">> HÃY MỞ APP ESPTouch v2 TRÊN ĐIỆN THOẠI!");
      ESP_LOGI(TAG,
               "============================================================");

      esp_smartconfig_set_type(SC_TYPE_ESPTOUCH);
      smartconfig_start_config_t cfg = SMARTCONFIG_START_CONFIG_DEFAULT();
      esp_smartconfig_start(&cfg);
    }
  } else if (base == IP_EVENT && id == IP_EVENT_STA_GOT_IP) {
    ip_event_got_ip_t *event = (ip_event_got_ip_t *)data;
    ESP_LOGI(TAG, "WiFi Connected! IP: " IPSTR, IP2STR(&event->ip_info.ip));
    s_retry_num = 0; // Đặt lại biến đếm khi đã có IP

    // Khi có mạng rồi, nối thẳng lên EMQX Cloud
    resolve_server_and_start_mqtt();
  }
  // 2. Nhóm sự kiện của hệ thống SmartConfig (thay cho Captive Portal)
  else if (base == SC_EVENT && id == SC_EVENT_GOT_SSID_PSWD) {
    ESP_LOGI(TAG, ">> Đã nhặt được mật khẩu WiFi từ App ESPTouch v2!");

    smartconfig_event_got_ssid_pswd_t *evt =
        (smartconfig_event_got_ssid_pswd_t *)data;
    wifi_config_t wifi_config;
    bzero(&wifi_config, sizeof(wifi_config_t));
    memcpy(wifi_config.sta.ssid, evt->ssid, sizeof(wifi_config.sta.ssid));
    memcpy(wifi_config.sta.password, evt->password,
           sizeof(wifi_config.sta.password));

    // Ngắt kết nối cũ và nạp cấu hình mới vào não
    esp_wifi_disconnect();
    esp_wifi_set_config(WIFI_IF_STA, &wifi_config);
    esp_wifi_connect();
  } else if (base == SC_EVENT && id == SC_EVENT_SEND_ACK_DONE) {
    ESP_LOGI(TAG, ">> Cấu hình hoàn tất! Đã báo tin về cho điện thoại.");
    esp_smartconfig_stop(); // Xong việc thì tắt chế độ dò tìm đi cho nhẹ RAM
  }
}

/**
 * @brief Thiết lập phần cứng WiFi (Antenna, bộ nhớ đệm, ngăn xếp giao thức)
 * VÀ Kích hoạt Event cho SmartConfig
 */
void wifi_init_sta(void) {
  // 1. Khởi tạo TCP/IP stack (cần thiết để gói tin có thể đi ra internet)
  ESP_ERROR_CHECK(esp_netif_init());

  // 2. Tạo vòng lặp sự kiện mặc định (nơi event_handler đăng ký lắng nghe)
  ESP_ERROR_CHECK(esp_event_loop_create_default());

  // 3. Cấu hình ESP32 thành chế độ Station (thiết bị khách nhận WiFi)
  // LƯU Ý: Không cần tạo AP (Phát WiFi) nữa vì SmartConfig chỉ cần lắng nghe
  // sóng
  esp_netif_create_default_wifi_sta();

  // 4. Đăng ký toàn bộ event cho event_handler
  ESP_ERROR_CHECK(esp_event_handler_register(WIFI_EVENT, ESP_EVENT_ANY_ID,
                                             &event_handler, NULL));
  ESP_ERROR_CHECK(esp_event_handler_register(IP_EVENT, IP_EVENT_STA_GOT_IP,
                                             &event_handler, NULL));
  ESP_ERROR_CHECK(
      esp_event_handler_register(SC_EVENT, ESP_EVENT_ANY_ID, &event_handler,
                                 NULL)); // Thêm event của SmartConfig

  // 5. Khởi tạo tài nguyên phần cứng WiFi (RAM, Buffer)
  wifi_init_config_t cfg = WIFI_INIT_CONFIG_DEFAULT();
  ESP_ERROR_CHECK(esp_wifi_init(&cfg));

  // 6. Bật WiFi và kết nối (Sẽ tự động nhảy vào SmartConfig nếu fail)
  ESP_ERROR_CHECK(esp_wifi_set_mode(WIFI_MODE_STA));
  ESP_ERROR_CHECK(esp_wifi_start());
}

// 4. MODULE CẢM BIẾN & HIỂN THỊ (LOGIC CHÍNH)
/**
 * @brief Thiết lập các thanh ghi cho bộ I2C Master
 */
static esp_err_t i2c_master_init(void) {
  i2c_config_t conf = {
      .mode = I2C_MODE_MASTER,
      .sda_io_num = I2C_MASTER_SDA_IO,
      .scl_io_num = I2C_MASTER_SCL_IO,
      .sda_pullup_en = GPIO_PULLUP_ENABLE, // Kích hoạt điện trở kéo lên cho SDA
      .scl_pullup_en = GPIO_PULLUP_ENABLE, // Kích hoạt điện trở kéo lên cho SCL
      .master.clk_speed = I2C_MASTER_FREQ_HZ,
  };
  i2c_param_config(I2C_MASTER_NUM, &conf);
  // Cài đặt driver vào hệ thống để bắt đầu truyền nhận
  return i2c_driver_install(I2C_MASTER_NUM, conf.mode, 0, 0, 0);
}

/**
 * @brief Task chạy song song: Vừa cập nhật OLED vừa bắn dữ liệu lên MQTT
 */
void sensor_display_task(void *pvParameters) {
  float temp = 0.0, hum = 0.0;
  char buffer[32];

  // Vẽ giao diện cố định (Static UI): Các chữ tiêu đề và Icon không thay đổi
  oled_draw_string(I2C_MASTER_NUM, 0, 0, "Temperature:");
  oled_draw_bitmap(I2C_MASTER_NUM, 96, 0, 32, 32, icon_temp_32x32);
  oled_draw_string(I2C_MASTER_NUM, 0, 4, "Humidity:");
  oled_draw_bitmap(I2C_MASTER_NUM, 96, 4, 32, 32, icon_hum_32x32);

  while (1) {
    // Đọc giá trị từ SHT30 qua I2C
    if (sht30_read(I2C_MASTER_NUM, &temp, &hum) == ESP_OK) {

      // --- PHẦN 1: CẬP NHẬT OLED ---
      sprintf(buffer, "%5.2f", temp);
      oled_draw_medium_numbers(I2C_MASTER_NUM, 0, 1, buffer);
      oled_draw_string(I2C_MASTER_NUM, 75, 1, "o");
      oled_draw_string(I2C_MASTER_NUM, 82, 2, "C");

      sprintf(buffer, "%5.2f", hum);
      oled_draw_medium_numbers(I2C_MASTER_NUM, 0, 5, buffer);
      oled_draw_string(I2C_MASTER_NUM, 80, 6, "%");

      // --- PHẦN 2: GỬI MQTT (Dữ liệu lên Web) ---
      if (mqtt_client != NULL) { // Chỉ gửi khi MQTT đã kết nối thành công
        char json_payload[128];
        // Tạo chuỗi JSON: {"temp": 32.50, "hum": 65.00}
        sprintf(json_payload, "{\"temp\": %.2f, \"hum\": %.2f}", temp, hum);

        // Publish lên Topic đồ án CE103, QoS level 1 (đảm bảo tới nơi ít nhất 1
        // lần)
        esp_mqtt_client_publish(mqtt_client,
                                "uit/ce103/project/nhietdovadoamphong",
                                json_payload, 0, 1, 0);
      }

      // In log ra Terminal máy tính để debug
      ESP_LOGI(TAG, "Dữ liệu: %.2f°C | %.2f%% | MQTT: %s", temp, hum,
               (mqtt_client ? "Connected" : "Waiting..."));

    } else {
      // Nếu hỏng dây I2C hoặc cảm biến chết thì báo lỗi lên màn hình
      ESP_LOGE(TAG, "Lỗi đọc cảm biến SHT30!");
      oled_draw_string(I2C_MASTER_NUM, 0, 1, "Sensor Err!  ");
    }

    // Nghỉ 2 giây trước khi lặp lại (giảm tải cho CPU và mạng)
    vTaskDelay(pdMS_TO_TICKS(2000));
  }
}

// 5. ĐIỂM BẮT ĐẦU CHƯƠNG TRÌNH (ENTRY POINT)
void app_main(void) {
  // 1. NVS Flash: WiFi cần bộ nhớ này để lưu các thông số hiệu chuẩn sóng vô
  // tuyến (RF). SmartConfig cũng lưu Pass vào đây để lần sau có điện là tự nhận
  // mạng.
  esp_err_t ret = nvs_flash_init();
  if (ret == ESP_ERR_NVS_NO_FREE_PAGES ||
      ret == ESP_ERR_NVS_NEW_VERSION_FOUND) {
    ESP_ERROR_CHECK(
        nvs_flash_erase()); // Nếu lỗi bộ nhớ thì xóa đi khởi tạo lại
    ret = nvs_flash_init();
  }
  ESP_ERROR_CHECK(ret);

  // 2. Khởi tạo I2C và màn hình OLED đầu tiên
  ESP_ERROR_CHECK(i2c_master_init());
  oled_init(I2C_MASTER_NUM);
  oled_clear(I2C_MASTER_NUM);

  // 3. Khởi tạo cấu hình WiFi (Sẽ tự động gọi SmartConfig nếu NVS trống)
  wifi_init_sta();

  // 4. Tạo một Task (Tiến trình) chạy độc lập để xử lý cảm biến và hiển thị
  // "sensor_task": Tên task
  // 4096: Kích thước vùng nhớ (Stack size) cấp cho task này
  // 5: Độ ưu tiên (Priority) - Càng cao càng quan trọng
  xTaskCreate(sensor_display_task, "sensor_task", 4096, NULL, 5, NULL);
}
