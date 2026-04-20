package main

import (
	"log"
	"math"
	"math/rand"
	"strconv"

	"sync"
	"time"

	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// Counter: Total de requisições HTTP
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "fiber_http_requests_total",
			Help: "Total de requisições HTTP recebidas pelo Fiber",
		},
		[]string{"method", "path", "status"},
	)

	// Histogram: Duração das requisições
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "fiber_http_request_duration_seconds",
			Help:    "Duração das requisições HTTP em segundos",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.0},
		},
		[]string{"method", "path", "status"},
	)

	// Gauge: Conexões ativas atuais
	activeConnections = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "fiber_active_connections",
			Help: "Número atual de conexões ativas no Fiber",
		},
	)

	// Gauge: Requisições em andamento
	requestsInFlight = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "fiber_requests_in_flight",
			Help: "Número de requisições sendo processadas atualmente",
		},
	)

	// Counter: Total de bytes processados
	httpBytesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "fiber_http_bytes_total",
			Help: "Total de bytes transferidos",
		},
		[]string{"direction"}, // "sent" ou "received"
	)

	// Simulação de conexões para demonstração
	currentConnCount float64 = 15
	connMutex        sync.Mutex
)

func init() {
	// Registra todas as métricas no Prometheus
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)
	prometheus.MustRegister(activeConnections)
	prometheus.MustRegister(requestsInFlight)
	prometheus.MustRegister(httpBytesTotal)
}

// Middleware customizado para instrumentação Prometheus
func prometheusMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.Path() == "/metrics" {
			return c.Next()
		}

		start := time.Now()

		requestsInFlight.Inc()
		defer requestsInFlight.Dec()

		err := c.Next()

		duration := time.Since(start).Seconds()

		method := c.Method()
		path := c.Route().Path
		if path == "" {
			path = c.Path()
		}
		status := strconv.Itoa(c.Response().StatusCode())

		// Registra as métricas
		httpRequestsTotal.WithLabelValues(method, path, status).Inc()
		httpRequestDuration.WithLabelValues(method, path, status).Observe(duration)

		responseSize := len(c.Response().Body())
		requestSize := len(c.Request().Body())
		httpBytesTotal.WithLabelValues("sent").Add(float64(responseSize))
		httpBytesTotal.WithLabelValues("received").Add(float64(requestSize))

		return err
	}
}

func main() {
	app := fiber.New(fiber.Config{
		AppName:      "Fiber Prometheus App v1.0.0",
		ServerHeader: "Fiber",
	})

	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${method} ${path} - ${latency}\n",
	}))
	app.Use(prometheusMiddleware())

	go simulateBackgroundActivity()

	setupRoutes(app)

	log.Println("Starting Fiber app on :8080")
	log.Fatal(app.Listen(":8080"))
}

func setupRoutes(app *fiber.App) {
	// Endpoint de saúde
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":    "healthy",
			"timestamp": time.Now().Unix(),
			"service":   "fiber-prometheus-app",
		})
	})

	app.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))

	users := app.Group("/api/v1/users")
	users.Get("/", getUsers)
	users.Post("/", createUser)
	users.Get("/:id", getUserByID)
	users.Put("/:id", updateUser)
	users.Delete("/:id", deleteUser)

	// Endpoint para simular carga
	app.Post("/load", func(c *fiber.Ctx) error {
		// Simula processamento com latência variável
		processingTime := time.Duration(rand.Intn(500)+50) * time.Millisecond
		time.Sleep(processingTime)

		// Simula falha ocasional
		if rand.Intn(100) < 5 { // 5% de chance de erro
			return c.Status(500).JSON(fiber.Map{
				"error": "Internal server error",
			})
		}

		return c.JSON(fiber.Map{
			"message":         "Load processed successfully",
			"processing_time": processingTime.String(),
		})
	})

	// Endpoint para informações de conexões
	app.Get("/connections", func(c *fiber.Ctx) error {
		connMutex.Lock()
		activeConn := int(math.Max(5, currentConnCount+float64(rand.Intn(10)-5)))
		connMutex.Unlock()

		return c.JSON(fiber.Map{
			"active_connections": activeConn,
			"max_connections":    100,
			"timestamp":          time.Now().Unix(),
		})
	})
}

// Handlers simulados para API de usuários
func getUsers(c *fiber.Ctx) error {
	// Simula latência de banco de dados
	time.Sleep(time.Duration(rand.Intn(100)+10) * time.Millisecond)

	return c.JSON([]fiber.Map{
		{"id": 1, "name": "Alice", "email": "alice@example.com"},
		{"id": 2, "name": "Bob", "email": "bob@example.com"},
		{"id": 3, "name": "Charlie", "email": "charlie@example.com"},
	})
}

func createUser(c *fiber.Ctx) error {
	time.Sleep(time.Duration(rand.Intn(200)+20) * time.Millisecond)

	// Simula falha ocasional
	if rand.Intn(100) < 3 {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid user data",
		})
	}

	return c.Status(201).JSON(fiber.Map{
		"id":      rand.Intn(1000) + 100,
		"name":    "New User",
		"email":   "newuser@example.com",
		"created": time.Now().Unix(),
	})
}

func getUserByID(c *fiber.Ctx) error {
	id := c.Params("id")
	time.Sleep(time.Duration(rand.Intn(80)+5) * time.Millisecond)

	// Simula usuário não encontrado
	if rand.Intn(100) < 10 {
		return c.Status(404).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	return c.JSON(fiber.Map{
		"id":    id,
		"name":  "User " + id,
		"email": "user" + id + "@example.com",
	})
}

func updateUser(c *fiber.Ctx) error {
	id := c.Params("id")
	time.Sleep(time.Duration(rand.Intn(150)+30) * time.Millisecond)

	return c.JSON(fiber.Map{
		"id":      id,
		"name":    "Updated User",
		"email":   "updated@example.com",
		"updated": time.Now().Unix(),
	})
}

func deleteUser(c *fiber.Ctx) error {
	time.Sleep(time.Duration(rand.Intn(100)+20) * time.Millisecond)

	return c.Status(204).Send(nil)
}

func simulateBackgroundActivity() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// Simula mudanças nas conexões ativas
		connMutex.Lock()
		delta := float64(rand.Intn(10) - 5)
		newVal := currentConnCount + delta
		if newVal < 5 {
			newVal = 5
		}
		if newVal > 150 {
			newVal = 150
		}
		currentConnCount = newVal
		connMutex.Unlock()
		activeConnections.Set(newVal)

		// Simula algumas requisições de background
		for i := 0; i < rand.Intn(5)+2; i++ {
			status := "200"
			path := "/background"
			method := "GET"

			if rand.Intn(100) > 90 {
				status = "500"
			}

			httpRequestsTotal.WithLabelValues(method, path, status).Inc()
			httpRequestDuration.WithLabelValues(method, path, status).Observe(rand.Float64() * 0.5)
		}
	}
}
