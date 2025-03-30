package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

func LoadConfig(path string) (Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return Config{}, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	var cfg Config
	sectionType := ""
	sectionName := ""
	currentCheck := Check{}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			if sectionType == "check" && currentCheck.Name != "" {
				cfg.Checks = append(cfg.Checks, currentCheck)
				currentCheck = Check{}
			}

			inside := strings.TrimSuffix(strings.TrimPrefix(line, "["), "]")
			parts := strings.SplitN(inside, " ", 2)
			sectionType = parts[0]
			sectionName = strings.Trim(parts[1], "\"")

			if sectionType == "check" {
				currentCheck.Name = sectionName
			}

			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return cfg, fmt.Errorf("invalid line: %s", line)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch sectionType {
		case "check":
			switch key {
			case "type":
				ct := CheckType(value)
				if !ct.IsValid() {
					return cfg, fmt.Errorf("invalid check type: %q", value)
				}
				currentCheck.Type = ct
			case "target":
				currentCheck.Target = value
			case "interval":
				currentCheck.Interval, _ = time.ParseDuration(value)
			case "timeout":
				currentCheck.Timeout, _ = time.ParseDuration(value)
			case "expect_status":
				fmt.Sscanf(value, "%d", &currentCheck.ExpectStatus)
			case "keyword":
				currentCheck.Keyword = strings.Trim(value, "\"")
			}
		case "alert":
			if sectionName == "email" {
				switch key {
				case "enabled":
					cfg.Alerts.Email.Enabled = (value == "true")
				case "to":
					cfg.Alerts.Email.To = value
				case "from":
					cfg.Alerts.Email.From = value
				case "smtp":
					cfg.Alerts.Email.SMTP = value
				case "username":
					cfg.Alerts.Email.Username = value
				case "password":
					cfg.Alerts.Email.Password = value
				}
			}
		}
	}

	if sectionType == "check" && currentCheck.Name != "" {
		cfg.Checks = append(cfg.Checks, currentCheck)
	}

	return cfg, scanner.Err()
}
