package cmd

import (
	"fmt"
	"log"
	"os"
	"text/tabwriter"
	"time"

	"github.com/mikhailbot/uptime-monitor/internal/state"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show latest status of all monitors",
	Run: func(cmd *cobra.Command, args []string) {
		db, err := state.Init(dbPath) // reuse --db flag
		if err != nil {
			log.Fatalf("❌ Failed to open DB: %v", err)
		}
		defer db.Close()

		latest, err := db.LatestStatuses()
		if err != nil {
			log.Fatalf("❌ Failed to load statuses: %v", err)
		}

		w := tabwriter.NewWriter(os.Stdout, 2, 4, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tTYPE\tSTATUS\tMESSAGE\tTIMESTAMP")
		for _, row := range latest {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", row.Name, row.Type, row.Status, row.Message, row.Timestamp.Format(time.RFC822))
		}
		w.Flush()
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
