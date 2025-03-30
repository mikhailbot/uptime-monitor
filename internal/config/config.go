package config

import "time"

type CheckType string

const (
	CheckHTTP    CheckType = "http"
	CheckKeyword CheckType = "keyword"
)

func (t CheckType) IsValid() bool {
	switch t {
	case CheckHTTP, CheckKeyword:
		return true
	default:
		return false
	}
}

type Check struct {
	Name         string
	Type         CheckType
	Target       string
	Interval     time.Duration
	Timeout      time.Duration
	ExpectStatus int
	Keyword      string
}

type EmailAlert struct {
	Enabled  bool
	To       string
	From     string
	SMTP     string
	Username string
	Password string
}

type Alerts struct {
	Email EmailAlert
}

type Config struct {
	Checks []Check
	Alerts Alerts
}
