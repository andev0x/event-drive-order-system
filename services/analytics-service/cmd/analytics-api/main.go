package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/andev0x/analytics-service/internal/cache"
	"github.com/andev0x/analytics-service/internal/handler"
	"github.com/andev0x/analytics-service/internal/model"
	"github.com/andev0x/analytics-service/internal/mq"
	"github.com/andev0x/analytics-service/internal/repository"
	"github.com/andev0x/analytics-service/internal/service"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	log.Println("Starting Analytics Service...")

	// Load configuration from environment variables
	config := loadConfig()

	// Initialize database
	log.Println("Connecting to database...")
	db, err := repository.InitDB(config.DBHost, config.DBPort, config.DBUser, config.DBPassword, config.DBName)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()
	log.Println("Database connected successfully")

	// Initialize Redis
	log.Println("Connecting to Redis...")
	redisClient, err := cache.InitRedis(config.RedisHost, config.RedisPort)
	if err != nil {
		log.Fatalf("Failed to initialize Redis: %v", err)
	}
	defer redisClient.Close()
	log.Println("Redis connected successfully")

	// Create repository, cache, and service
	analyticsRepo := repository.NewMySQLAnalyticsRepository(db)
	analyticsCache := cache.NewRedisAnalyticsCache(redisClient)
	analyticsService := service.NewAnalyticsService(analyticsRepo, analyticsCache)

	// Create handler
	analyticsHandler := handler.NewAnalyticsHandler(analyticsService)

	// Initialize RabbitMQ consumer
	log.Println("Connecting to RabbitMQ...")
	consumer, err := mq.NewRabbitMQConsumer(config.RabbitMQURL)
	if err != nil {
		log.Fatalf("Failed to initialize RabbitMQ consumer: %v", err)
	}
	defer consumer.Close()
	log.Println("RabbitMQ connected successfully")

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
			return consumer.HealthCheck()
		},
	}
	analyticsHandler.SetHealthChecker(healthChecker)

	// Start consuming events
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = consumer.StartConsuming(ctx, func(event *model.OrderCreatedEvent) error {
		return analyticsService.ProcessOrderEvent(context.Background(), event)
	})
	if err != nil {
		log.Fatalf("Failed to start consuming: %v", err)
	}

	// Setup router
	router := mux.NewRouter()

	// Health check
	router.HandleFunc("/health", analyticsHandler.HealthCheck).Methods("GET")

	// Metrics endpoint
	router.Handle("/metrics", promhttp.Handler()).Methods("GET")

	// Analytics endpoints
	router.HandleFunc("/analytics/summary", analyticsHandler.GetSummary).Methods("GET")

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
		log.Printf("Analytics Service listening on port %s", config.ServicePort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	cancel() // Stop the consumer

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
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
		DBUser:      getEnv("DB_USER", "analyticsuser"),
		DBPassword:  getEnv("DB_PASSWORD", "analyticspass"),
		DBName:      getEnv("DB_NAME", "analytics_db"),
		RedisHost:   getEnv("REDIS_HOST", "localhost"),
		RedisPort:   getEnv("REDIS_PORT", "6379"),
		RabbitMQURL: getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		ServicePort: getEnv("SERVICE_PORT", "8081"),
	}
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
