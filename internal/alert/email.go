package alert

import (
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strings"
)

func SendEmailAlert(temp float64, level string) {
	from := os.Getenv("SMTP_FROM")
	password := os.Getenv("SMTP_PASSWORD")
	toRaw := os.Getenv("SMTP_TO")

	if from == "" || password == "" || toRaw == "" {
		log.Println("[ALERT] Missing SMTP credentials in .env")
		return
	}

	to := strings.Split(toRaw, ",")
	for i := range to {
		to[i] = strings.TrimSpace(to[i])
	}

	var subject string
	if level == "critical" {
		subject = "[CRITICAL] DANGEROUS Temperature Level!"
	} else {
		subject = "[WARNING] High Room Temperature Alert"
	}

	body := fmt.Sprintf("Current temperature is %.2f°C.\nPlease check the room immediately!", temp)
	msg := []byte(fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s\r\n", toRaw, subject, body))

	auth := smtp.PlainAuth("", from, password, "smtp.gmail.com")
	err := smtp.SendMail("smtp.gmail.com:587", auth, from, to, msg)
	if err != nil {
		log.Printf("[ALERT] Error sending email: %v\n", err)
		return
	}

	log.Printf("[ALERT] Alert email (%s) sent successfully to %v\n", level, to)
}
