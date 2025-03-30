package monitor

import (
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/mikhailbot/uptime-monitor/internal/alert"
	"github.com/mikhailbot/uptime-monitor/internal/config"
	"github.com/mikhailbot/uptime-monitor/internal/state"
)

var checkState = make(map[string]*checkTracker)

type checkTracker struct {
	lastStatus string
	streak     int
	alerted    bool
}

func Run(cfg config.Config, db *state.DB, a config.Alerts) {
	for _, check := range cfg.Checks {
		go superviseCheck(check, db, a)
	}
	select {} // keep the main goroutine alive
}

func superviseCheck(c config.Check, db *state.DB, a config.Alerts) {
	for {
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("🔥 Check %q crashed: %v — restarting in 15s", c.Name, r)
					time.Sleep(15 * time.Second)
				}
			}()
			runCheckLoop(c, db, a)
		}()
	}
}

func runCheckLoop(c config.Check, db *state.DB, a config.Alerts) {
	log.Printf("✅ Starting check loop: %s (%s)", c.Name, c.Target)

	ticker := time.NewTicker(c.Interval)
	defer ticker.Stop()

	key := createKey(c.Name, c.Type)
	if _, exists := checkState[key]; !exists {
		checkState[key] = &checkTracker{}
	}

	for {
		err := runCheck(c)
		status := "ok"
		message := "success"
		if err != nil {
			status = "fail"
			message = err.Error()
		}

		db.SaveResult(state.CheckResult{
			Name:      key,
			Type:      string(c.Type),
			Status:    status,
			Message:   message,
			Timestamp: time.Now(),
		})

		// Check state and alert if appropriate
		tracker := checkState[key]
		if status == tracker.lastStatus {
			tracker.streak++
		} else {
			tracker.streak = 1
		}
		tracker.lastStatus = status

		if status == "fail" && tracker.streak == 3 && !tracker.alerted {
			log.Printf("🔴 ALERT: %s is DOWN (3 consecutive failures)", c.Name)
			tracker.alerted = true

			alert.SendAll(a, alert.DownPayload(c.Name, c.Target, string(c.Type)))

		}

		if status == "ok" && tracker.alerted && tracker.streak == 3 {
			log.Printf("🟢 RECOVERY: %s is UP (3 consecutive successes)", c.Name)
			tracker.alerted = false

			alert.SendAll(a, alert.UpPayload(c.Name, c.Target, string(c.Type)))

		}

		<-ticker.C
	}
}

func runCheck(c config.Check) error {
	client := &http.Client{
		Timeout: c.Timeout,
	}

	resp, err := client.Get(c.Target)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if c.ExpectStatus != 0 && resp.StatusCode != c.ExpectStatus {
		return &CheckError{"unexpected status", resp.StatusCode}
	}

	if c.Type == config.CheckKeyword && c.Keyword != "" {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		if !strings.Contains(string(body), c.Keyword) {
			return &CheckError{"keyword not found", 0}
		}
	}

	return nil
}

func createKey(name string, typ config.CheckType) string {
	var invalidChars = regexp.MustCompile(`[^a-zA-Z0-9_-]+`)

	safeName := strings.ToLower(strings.TrimSpace(name))
	safeName = invalidChars.ReplaceAllString(safeName, "-")

	return safeName + "_" + string(typ)
}

type CheckError struct {
	Reason string
	Status int
}

func (e *CheckError) Error() string {
	if e.Status != 0 {
		return e.Reason + ": HTTP " + http.StatusText(e.Status)
	}
	return e.Reason
}
