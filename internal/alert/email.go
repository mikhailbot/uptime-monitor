package alert

import (
	"fmt"
	"log"
	"net"
	"net/smtp"
	"time"

	"github.com/mikhailbot/uptime-monitor/internal/config"
)

const emailTimeout = 10 * time.Second

func SendEmail(cfg config.EmailAlert, subject, body string) error {
	if !cfg.Enabled {
		return nil
	}

	auth := smtp.PlainAuth("", cfg.Username, cfg.Password, smtpHost(cfg.SMTP))
	msg := []byte(fmt.Sprintf(
		"To: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		cfg.To, subject, body,
	))

	log.Printf("📧 Sending email to %s via %s", cfg.To, cfg.SMTP)

	result := make(chan error, 1)
	go func() {
		result <- smtp.SendMail(cfg.SMTP, auth, cfg.From, []string{cfg.To}, msg)
	}()

	select {
	case err := <-result:
		if err != nil {
			log.Printf("❌ Failed to send email: %v", err)
		} else {
			log.Println("✅ Email sent successfully")
		}
		return err
	case <-time.After(emailTimeout):
		log.Printf("⏱️ Email send timed out after %s", emailTimeout)
		return fmt.Errorf("email send timeout after %s", emailTimeout)
	}
}

func smtpHost(addr string) string {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		log.Printf("⚠️ Failed to parse SMTP host: %v", err)
		return addr
	}
	return host
}
