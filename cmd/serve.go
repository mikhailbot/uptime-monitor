package cmd

import (
	"log"

	"github.com/mikhailbot/uptime-monitor/internal/config"
	"github.com/mikhailbot/uptime-monitor/internal/monitor"
	"github.com/mikhailbot/uptime-monitor/internal/state"
	"github.com/spf13/cobra"
)

var (
	configFile string
	dbPath     string
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run the monitoring daemon",
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("🔧 Starting uptime monitor daemon")
		log.Printf("📄 Using config file: %s", configFile)

		cfg, err := config.LoadConfig(configFile)
		if err != nil {
			log.Fatalf("❌ Error loading config: %v", err)
		}

		log.Printf("✅ Loaded %d checks", len(cfg.Checks))
		for _, check := range cfg.Checks {
			switch check.Type {
			case config.CheckHTTP:
				log.Printf("   ↪️  %s [http]    %s every %s", check.Name, check.Target, check.Interval)
			case config.CheckKeyword:
				log.Printf("   ↪️  %s [keyword] %s every %s (looking for: %q)", check.Name, check.Target, check.Interval, check.Keyword)
			default:
				log.Printf("   ⚠️  %s [unknown type: %s]", check.Name, check.Type)
			}
		}
		if cfg.Alerts.Email.Enabled {
			log.Printf("📧 Email alerts enabled → %s via %s", cfg.Alerts.Email.To, cfg.Alerts.Email.SMTP)
		} else {
			log.Printf("📧 Email alerts disabled")
		}

		log.Printf("📦 Using SQLite database: %s", dbPath)
		db, err := state.Init(dbPath)
		if err != nil {
			log.Fatalf("❌ Error initializing database: %v", err)
		}
		defer db.Close()

		monitor.Run(cfg, db, cfg.Alerts)
	},
}

func init() {
	serveCmd.Flags().StringVarP(&configFile, "config", "c", "config.ini", "Path to config file")
	serveCmd.Flags().StringVar(&dbPath, "db", "monitor.db", "Path to SQLite database file")
	rootCmd.AddCommand(serveCmd)
}
