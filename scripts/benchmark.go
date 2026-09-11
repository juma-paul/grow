//go:build ignore

// benchmark.go — sweep concurrency levels and record latency + CPU.
//
// Usage:
//   go run scripts/benchmark.go -addr localhost:8080
//
// Runs the /execute endpoint at concurrency levels 1, 5, 10, 25, 50, 100
// and prints a CSV table of p50/p95/p99 latency per level.
// If the server exposes /metrics (Prometheus), also captures
// process_cpu_seconds_total deltas.

package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

var levels = []int{1, 5, 10, 25, 50, 100}

type result struct {
	conc     int
	p50      time.Duration
	p95      time.Duration
	p99      time.Duration
	avg      time.Duration
	rps      float64
	errors   int64
	cpuDelta float64
}

func main() {
	addr := flag.String("addr", "localhost:8080", "server address")
	count := flag.Int("count", 50, "elements per run")
	perLevel := flag.Int("n", 200, "requests per level")
	flag.Parse()

	fmt.Println("conc,p50_ms,p95_ms,p99_ms,avg_ms,rps,errors,cpu_delta_s")

	for _, c := range levels {
		r := runLevel(*addr, c, *count, *perLevel)
		fmt.Printf("%d,%.1f,%.1f,%.1f,%.1f,%.1f,%d,%.3f\n",
			r.conc,
			float64(r.p50.Microseconds())/1000.0,
			float64(r.p95.Microseconds())/1000.0,
			float64(r.p99.Microseconds())/1000.0,
			float64(r.avg.Microseconds())/1000.0,
			r.rps,
			r.errors,
			r.cpuDelta,
		)
		time.Sleep(500 * time.Millisecond)
	}
}

func runLevel(addr string, conc, count, n int) result {
	q := url.Values{}
	q.Set("count", fmt.Sprintf("%d", count))
	q.Set("strategy", "cpython")
	wsURL := fmt.Sprintf("ws://%s/execute?%s", addr, q.Encode())

	cpuBefore := getCPU(addr)

	var (
		mu        sync.Mutex
		latencies []time.Duration
		errors    int64
	)

	sem := make(chan struct{}, conc)
	var wg sync.WaitGroup
	start := time.Now()

	for i := 0; i < n; i++ {
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

			for {
				if _, _, err := conn.ReadMessage(); err != nil {
					break
				}
			}

			mu.Lock()
			latencies = append(latencies, time.Since(t0))
			mu.Unlock()
		}()
	}

	wg.Wait()
	elapsed := time.Since(start)

	cpuAfter := getCPU(addr)

	if len(latencies) == 0 {
		return result{conc: conc, errors: errors}
	}

	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })

	var sum time.Duration
	for _, l := range latencies {
		sum += l
	}

	return result{
		conc:     conc,
		p50:      latencies[len(latencies)*50/100],
		p95:      latencies[len(latencies)*95/100],
		p99:      latencies[len(latencies)*99/100],
		avg:      sum / time.Duration(len(latencies)),
		rps:      float64(len(latencies)) / elapsed.Seconds(),
		errors:   errors,
		cpuDelta: cpuAfter - cpuBefore,
	}
}

func getCPU(addr string) float64 {
	resp, err := http.Get(fmt.Sprintf("http://%s/metrics", addr))
	if err != nil {
		return 0
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var cpu float64
	for _, line := range splitLines(string(body)) {
		if len(line) > 0 && line[0] != '#' {
			if match("process_cpu_seconds_total", line) {
				fmt.Sscanf(line, "process_cpu_seconds_total %f", &cpu)
				return cpu
			}
		}
	}
	return 0
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func match(prefix, line string) bool {
	return len(line) >= len(prefix) && line[:len(prefix)] == prefix
}
