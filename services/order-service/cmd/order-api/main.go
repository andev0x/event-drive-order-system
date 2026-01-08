// Package main provides the entry point for the order service API.
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/andev0x/order-service/internal/cache"
	"github.com/andev0x/order-service/internal/handler"
	"github.com/andev0x/order-service/internal/mq"
	"github.com/andev0x/order-service/internal/repository"
	"github.com/andev0x/order-service/internal/service"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	log.Println("Starting Order Service...")

	// Load configuration from environment variables
	config := loadConfig()

	// Initialize database
	log.Println("Connecting to database...")
	db, err := repository.InitDB(config.DBHost, config.DBPort, config.DBUser, config.DBPassword, config.DBName)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database connection: %v", err)
		}
	}()
	log.Println("Database connected successfully")

	// Initialize Redis
	log.Println("Connecting to Redis...")
	redisClient, err := cache.InitRedis(config.RedisHost, config.RedisPort)
	if err != nil {
		log.Printf("Failed to initialize Redis: %v", err)
		return
	}
	defer func() {
		if err := redisClient.Close(); err != nil {
			log.Printf("Error closing Redis connection: %v", err)
		}
	}()
	log.Println("Redis connected successfully")

	// Initialize RabbitMQ publisher
	log.Println("Connecting to RabbitMQ...")
	publisher, err := mq.NewRabbitMQPublisher(config.RabbitMQURL)
	if err != nil {
		log.Printf("Failed to initialize RabbitMQ publisher: %v", err)
		return
	}
	defer func() {
		if err := publisher.Close(); err != nil {
			log.Printf("Error closing RabbitMQ publisher: %v", err)
		}
	}()
	log.Println("RabbitMQ connected successfully")

	// Create repository, cache, and service
	orderRepo := repository.NewMySQLOrderRepository(db)
	orderCache := cache.NewRedisOrderCache(redisClient)
	orderService := service.NewOrderService(orderRepo, orderCache, publisher)

	// Create handler
	orderHandler := handler.NewOrderHandler(orderService)

	// Setup health checker
	healthChecker := &handler.HealthChecker{
		DBHealthFunc: func() error {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			return db.PingContext(ctx)
		},
		CacheHealthFunc: func() error {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			return redisClient.Ping(ctx).Err()
		},
		MQHealthFunc: func() error {
			return publisher.HealthCheck()
		},
	}
	orderHandler.SetHealthChecker(healthChecker)

	// Setup router
	router := mux.NewRouter()

	// Health check
	router.HandleFunc("/health", orderHandler.HealthCheck).Methods("GET")

	// Metrics endpoint
	router.Handle("/metrics", promhttp.Handler()).Methods("GET")

	// Order endpoints
	router.HandleFunc("/orders", orderHandler.CreateOrder).Methods("POST")
	router.HandleFunc("/orders/{id}", orderHandler.GetOrder).Methods("GET")
	router.HandleFunc("/orders", orderHandler.ListOrders).Methods("GET")

	// Setup server
	srv := &http.Server{
		Addr:         ":" + config.ServicePort,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Order Service listening on port %s", config.ServicePort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}

// Config holds application configuration
type Config struct {
	DBHost      string
	DBPort      string
	DBUser      string
	DBPassword  string
	DBName      string
	RedisHost   string
	RedisPort   string
	RabbitMQURL string
	ServicePort string
}

// loadConfig loads configuration from environment variables
func loadConfig() Config {
	return Config{
		DBHost:      getEnv("DB_HOST", "localhost"),
		DBPort:      getEnv("DB_PORT", "3306"),
		DBUser:      getEnv("DB_USER", "orderuser"),
		DBPassword:  getEnv("DB_PASSWORD", "orderpass"),
		DBName:      getEnv("DB_NAME", "order_db"),
		RedisHost:   getEnv("REDIS_HOST", "localhost"),
		RedisPort:   getEnv("REDIS_PORT", "6379"),
		RabbitMQURL: getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		ServicePort: getEnv("SERVICE_PORT", "8080"),
	}
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
