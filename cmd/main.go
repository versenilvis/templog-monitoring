package main

import (
	"log"

	"github.com/gofiber/contrib/v3/websocket"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/joho/godotenv"
	"github.com/versenilvis/templog-monitoring/internal/db"
	"github.com/versenilvis/templog-monitoring/internal/hub"
	"github.com/versenilvis/templog-monitoring/internal/sensor"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("[INFO] No .env file found or error reading it, using environment variables")
	}

	db.Init()

	h := hub.New(60)

	// usb port version
	// go sensor.ReadUart(h, "/dev/ttyUSB0")

	// wifi version
	go sensor.ReadMQTT(h)

	app := fiber.New(fiber.Config{
		AppName: "templog",
	})

	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
	}))

	app.Use("/ws", func(c fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get("/ws", websocket.New(func(c *websocket.Conn) {
		h.Register(c)
		defer h.Unregister(c)

		for {
			if _, _, err := c.ReadMessage(); err != nil {
				break
			}
		}
	}))

	app.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	app.Get("/api/stats", func(c fiber.Ctx) error {
		stats, err := db.GetTodayStats()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(stats)
	})

	log.Fatal(app.Listen(":8080"))
}
