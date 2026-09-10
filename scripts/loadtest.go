// loadtest.go — concurrent WebSocket load-test harness for grow.
//
// Usage:
//   go run scripts/loadtest.go -endpoint /execute -count 20 -conc 10 -n 100
//
// Flags:
//   -addr     server address (default localhost:8080)
//   -endpoint one of /execute, /auto, /observe (default /execute)
//   -count    elements per run (default 20)
//   -strategy growth strategy (default cpython)
//   -conc     concurrent connections (default 10)
//   -n        total requests to make (default 100)
//   -source   Python source for /auto or /observe

package main

import (
	"flag"
	"fmt"
	"math"
	"net/url"
	"os"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

func main() {
	addr := flag.String("addr", "localhost:8080", "server address")
	endpoint := flag.String("endpoint", "/execute", "endpoint: /execute, /auto, /observe")
	count := flag.Int("count", 20, "elements per run")
	strategy := flag.String("strategy", "cpython", "growth strategy")
	conc := flag.Int("conc", 10, "concurrent connections")
	n := flag.Int("n", 100, "total requests")
	source := flag.String("source", "", "Python source for /auto or /observe")
	flag.Parse()

	defaultSource := "lst = []\nfor i in range(50):\n    lst.append(i)\n"
	if *source == "" {
		*source = defaultSource
	}

	q := url.Values{}
	q.Set("count", fmt.Sprintf("%d", *count))
	q.Set("strategy", *strategy)
	wsURL := fmt.Sprintf("ws://%s%s?%s", *addr, *endpoint, q.Encode())

	fmt.Printf("Load test: %s\n", wsURL)
	fmt.Printf("Concurrency: %d, Total: %d\n\n", *conc, *n)

	var (
		mu        sync.Mutex
		latencies []time.Duration
		errors    int64
		events    int64
	)

	sem := make(chan struct{}, *conc)
	var wg sync.WaitGroup

	start := time.Now()

	for i := 0; i < *n; i++ {
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()

			t0 := time.Now()
			conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			if err != nil {
				atomic.AddInt64(&errors, 1)
				return
			}
			defer conn.Close()

			needsSource := *endpoint == "/auto" || *endpoint == "/observe"
			if needsSource {
				if err := conn.WriteMessage(websocket.TextMessage, []byte(*source)); err != nil {
					atomic.AddInt64(&errors, 1)
					return
				}
			}

			var evtCount int64
			for {
				_, _, err := conn.ReadMessage()
				if err != nil {
					break
				}
				evtCount++
			}

			lat := time.Since(t0)
			atomic.AddInt64(&events, evtCount)

			mu.Lock()
			latencies = append(latencies, lat)
			mu.Unlock()
		}()
	}

	wg.Wait()
	elapsed := time.Since(start)

	if len(latencies) == 0 {
		fmt.Println("All requests failed!")
		os.Exit(1)
	}

	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })

	p50 := latencies[len(latencies)*50/100]
	p95 := latencies[len(latencies)*95/100]
	p99 := latencies[len(latencies)*99/100]

	var sum float64
	for _, l := range latencies {
		sum += l.Seconds()
	}
	avg := sum / float64(len(latencies))

	var variance float64
	for _, l := range latencies {
		d := l.Seconds() - avg
		variance += d * d
	}
	stddev := math.Sqrt(variance / float64(len(latencies)))

	rps := float64(len(latencies)) / elapsed.Seconds()

	fmt.Printf("Results (%d/%d succeeded):\n", len(latencies), *n)
	fmt.Printf("  Elapsed:   %s\n", elapsed.Round(time.Millisecond))
	fmt.Printf("  RPS:       %.1f\n", rps)
	fmt.Printf("  Events:    %d total\n", events)
	fmt.Printf("  Errors:    %d\n", errors)
	fmt.Println()
	fmt.Printf("  Avg:       %s\n", time.Duration(avg*float64(time.Second)).Round(time.Microsecond))
	fmt.Printf("  Stddev:    %s\n", time.Duration(stddev*float64(time.Second)).Round(time.Microsecond))
	fmt.Printf("  p50:       %s\n", p50.Round(time.Microsecond))
	fmt.Printf("  p95:       %s\n", p95.Round(time.Microsecond))
	fmt.Printf("  p99:       %s\n", p99.Round(time.Microsecond))
	fmt.Printf("  Min:       %s\n", latencies[0].Round(time.Microsecond))
	fmt.Printf("  Max:       %s\n", latencies[len(latencies)-1].Round(time.Microsecond))
}
