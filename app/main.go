package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// Counter: Total de requisições
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total de requisições HTTP recebidas",
		},
		[]string{"method", "status"},
	)

	// Gauge: Conexões ativas
	activeConnections = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_connections",
			Help: "Número de conexões ativas",
		},
	)

	// Histogram: Latência de requisições
	requestDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "request_duration_seconds",
			Help:    "Duração das requisições em segundos",
			Buckets: []float64{0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0},
		},
	)

	// Track current connection count for handleConnections endpoint
	currentConnCount float64 = 10
	connMutex        sync.Mutex
)

func init() {
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(activeConnections)
	prometheus.MustRegister(requestDuration)
}

func main() {
	// Background goroutine que simula atividade
	go simulateActivity()

	// Handlers
	http.HandleFunc("/request/", handleRequest)
	http.HandleFunc("/health", handleHealth)
	http.HandleFunc("/connections", handleConnections)
	http.Handle("/metrics", promhttp.Handler())

	log.Println("Starting app on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		httpRequestsTotal.WithLabelValues("POST", "405").Inc()
		return
	}

	start := time.Now()

	// Simula latência aleatória entre 10ms e 1s
	latency := time.Duration(rand.Intn(990)+10) * time.Millisecond
	time.Sleep(latency)

	duration := time.Since(start).Seconds()
	requestDuration.Observe(duration)

	w.WriteHeader(http.StatusOK)
	httpRequestsTotal.WithLabelValues("POST", "200").Inc()
	fmt.Fprintf(w, "OK\n")
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK\n")
	httpRequestsTotal.WithLabelValues("GET", "200").Inc()
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	connMutex.Lock()
	activeConn := int(math.Max(2, currentConnCount+float64(rand.Intn(5)-2)))
	connMutex.Unlock()
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "{\"active_connections\": %d}\n", activeConn)
	httpRequestsTotal.WithLabelValues("GET", "200").Inc()
}

func simulateActivity() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// Incrementa requisições aleatoriamente
		for i := 0; i < rand.Intn(10)+5; i++ {
			status := "200"
			if rand.Intn(100) > 95 {
				status = "500"
			}
			httpRequestsTotal.WithLabelValues("GET", status).Inc()
		}

		// Incrementa conexões ativas aleatoriamente
		delta := float64(rand.Intn(5) - 2)
		connMutex.Lock()
		newVal := currentConnCount + delta
		if newVal < 1 {
			newVal = 1
		}
		if newVal > 100 {
			newVal = 100
		}
		currentConnCount = newVal
		connMutex.Unlock()
		activeConnections.Set(newVal)

		// Simula histograma
		latency := time.Duration(rand.Intn(900)+100) * time.Millisecond
		requestDuration.Observe(latency.Seconds())
	}
}
