package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

var (
	mu               sync.Mutex
	failStreak       = 0
	mustSucceedCount = 0
	failChance       = 0.1
	failBurstLen     = 5
	recoveryStreak   = 3
)

func main() {
	rand.Seed(time.Now().UnixNano())

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()

		switch {
		// 🔁 Recovery mode: must send 3 good responses
		case mustSucceedCount > 0:
			mustSucceedCount--
			log.Printf("🟢 Forced success during recovery (%d left)", mustSucceedCount)
			writeOK(w)

		// 🔥 Failure mode
		case failStreak > 0:
			failStreak--
			log.Printf("🔴 Simulated outage (%d failures left)", failStreak)
			http.Error(w, "💥 simulated outage", http.StatusInternalServerError)

			// When failure mode ends, begin recovery
			if failStreak == 0 {
				mustSucceedCount = recoveryStreak
				log.Println("🔁 Entering recovery mode")
			}

		// 🎲 Normal mode: maybe trigger failure
		default:
			if rand.Float64() < failChance {
				failStreak = failBurstLen
				log.Printf("⚠️  Entering failure mode (%d failures)", failBurstLen)
				http.Error(w, "💥 simulated outage", http.StatusInternalServerError)
				failStreak-- // count this request as the first failure
			} else {
				writeOK(w)
			}
		}
	})

	port := 8080
	log.Printf("🚀 Starting test server on :%d", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

func writeOK(w http.ResponseWriter) {
	log.Println("✅ Responded with 200 OK")
	fmt.Fprintln(w, "✅ Hello, monitor — keyword: RD")
}
