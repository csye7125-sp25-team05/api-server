package main

import (
	"api-server/internal/api"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	requestCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "api_server_requests_total",
			Help: "Total number of requests to the API server by endpoint",
		},
		[]string{"endpoint", "method"},
	)
	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "api_server_request_duration_seconds",
			Help:    "Duration of API server requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"endpoint", "method"},
	)
	goGoroutines = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "api_server_go_goroutines",
			Help: "Number of goroutines in the Go runtime",
		},
	)
	goMemory = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "api_server_go_memory_bytes",
			Help: "Amount of memory allocated by the Go runtime in bytes",
		},
	)
)

func init() {
	// Register Prometheus metrics
	prometheus.MustRegister(requestCounter)
	prometheus.MustRegister(requestDuration)
	prometheus.MustRegister(goGoroutines)
	prometheus.MustRegister(goMemory)

	// Start a goroutine to update Go runtime metrics periodically
	go func() {
		for {
			var memStats runtime.MemStats
			runtime.ReadMemStats(&memStats)
			goGoroutines.Set(float64(runtime.NumGoroutine()))
			goMemory.Set(float64(memStats.Alloc))
			time.Sleep(30 * time.Second) // Update every 30 seconds
		}
	}()
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: Error loading .env file, relying on environment variables:", err)
	}

	fmt.Println("Starting API server...")
	// db, err := sql.Open("postgres", "user=postgres dbname=api options=-csearch_path=api,public sslmode=disable")

	// Get database connection details from environment variables
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	// Construct the connection string
	connStr := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=disable options=-csearch_path=api,public",
		dbUser, dbPassword, dbHost, dbPort, dbName)
	fmt.Println("🔌 DB Connection String:", connStr)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to open database connection: %v", err)
	}
	defer db.Close()

	// Test database connection
	fmt.Println("Testing database connection...")
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	fmt.Println("Database connection successful")

	// Set up router with middleware for metrics
	router := mux.NewRouter()

	// Middleware to record request metrics
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			endpoint := r.URL.Path
			method := r.Method

			// Increment request counter
			requestCounter.WithLabelValues(endpoint, method).Inc()

			// Call the next handler
			next.ServeHTTP(w, r)

			// Record request duration
			duration := time.Since(start).Seconds()
			requestDuration.WithLabelValues(endpoint, method).Observe(duration)
		})
	})

	// Expose Prometheus metrics endpoint first
	router.Handle("/metrics", promhttp.Handler()).Methods("GET")

	// Add application routes
	appRouter := api.SetupRoutes(db)
	router.PathPrefix("/").Handler(appRouter)

	// Start server
	log.Fatal(http.ListenAndServe(":8080", router))
}
