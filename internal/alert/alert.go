package alert

import (
	"fmt"

	"github.com/mikhailbot/uptime-monitor/internal/config"
)

type AlertType string

const (
	AlertEmail AlertType = "email"
)

type AlertPayload struct {
	Subject string
	Body    string
}

func DownPayload(checkName, target string, checkType string) AlertPayload {
	return AlertPayload{
		Subject: fmt.Sprintf("ALERT: %s", checkName),
		Body: fmt.Sprintf(`
			<h2 style="color: #c10007;">%s is currenly unavailable</h2>
			<p>The check for <strong>%s</strong> has failed <strong>3 times in a row</strong>.</p>
			<p>Target: <a href="%s">%s</a><br>
			Alert Type: <strong>%s</strong></p>
			<br>
			<p>Uptime Monitor</p>
		`, checkName, checkName, target, target, checkType),
	}
}

func UpPayload(checkName, target string, checkType string) AlertPayload {
	return AlertPayload{
		Subject: fmt.Sprintf("RECOVERED: %s", checkName),
		Body: fmt.Sprintf(`
			<h2 style="color: #00a63e;">%s is back online</h2>
			<p>The check for <strong>%s</strong> has recovered after <strong>3 consecutive successes</strong>.</p>
			<p>Target: <a href="%s">%s</a><br>
			Alert Type: <strong>%s</strong></p>
			<br>
			<p>Uptime Monitor</p>
		`, checkName, checkName, target, target, checkType),
	}
}

func SendAll(cfg config.Alerts, payload AlertPayload) {
	if cfg.Email.Enabled {
		go SendEmail(cfg.Email, payload.Subject, payload.Body)
	}
}
