// Package main implements the notification worker service that consumes order events from RabbitMQ
// and sends notifications to customers.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	exchangeName = "orders"
	exchangeType = "topic"
	queueName    = "notifications.orders"
	routingKey   = "order.created"
)

// OrderCreatedEvent represents the event consumed from RabbitMQ
type OrderCreatedEvent struct {
	OrderID     string    `json:"order_id"`
	CustomerID  string    `json:"customer_id"`
	ProductID   string    `json:"product_id"`
	Quantity    int       `json:"quantity"`
	TotalAmount float64   `json:"total_amount"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	EventType   string    `json:"event_type"`
}

func main() {
	log.Println("Starting Notification Worker...")

	// Get RabbitMQ URL from environment
	rabbitMQURL := getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")
	healthPort := getEnv("HEALTH_PORT", "8082")

	// Connect to RabbitMQ
	conn, err := connectRabbitMQ(rabbitMQURL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("Error closing RabbitMQ connection: %v", err)
		}
	}()

	// Start health check HTTP server
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		healthCheck(w, r, conn)
	})

	server := &http.Server{
		Addr:         ":" + healthPort,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Printf("Health check server listening on port %s", healthPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Health check server error: %v", err)
		}
	}()

	// Create channel
	channel, err := conn.Channel()
	if err != nil {
		log.Printf("Failed to open channel: %v", err)
		return
	}
	defer func() {
		if err := channel.Close(); err != nil {
			log.Printf("Error closing RabbitMQ channel: %v", err)
		}
	}()

	// Declare exchange
	err = channel.ExchangeDeclare(
		exchangeName,
		exchangeType,
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		log.Printf("Failed to declare exchange: %v", err)
		return
	}

	// Declare queue
	queue, err := channel.QueueDeclare(
		queueName,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		log.Printf("Failed to declare queue: %v", err)
		return
	}

	// Bind queue to exchange
	err = channel.QueueBind(
		queue.Name,
		routingKey,
		exchangeName,
		false,
		nil,
	)
	if err != nil {
		log.Printf("Failed to bind queue: %v", err)
		return
	}

	log.Printf("Notification worker connected to queue '%s'", queueName)

	// Set QoS
	err = channel.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		log.Printf("Failed to set QoS: %v", err)
		return
	}

	// Start consuming
	msgs, err := channel.Consume(
		queue.Name,
		"",    // consumer
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		log.Printf("Failed to register consumer: %v", err)
		return
	}

	log.Println("Notification worker is now consuming order events...")

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("Shutting down notification worker...")
		cancel()
	}()

	// Process messages
	for {
		select {
		case <-ctx.Done():
			log.Println("Notification worker stopped")
			return
		case msg, ok := <-msgs:
			if !ok {
				log.Println("Message channel closed")
				return
			}

			// Parse event
			var event OrderCreatedEvent
			if err := json.Unmarshal(msg.Body, &event); err != nil {
				log.Printf("Error unmarshaling event: %v", err)
				if nackErr := msg.Nack(false, false); nackErr != nil {
					log.Printf("Error nacking message: %v", nackErr)
				}
				continue
			}

			log.Printf("Received OrderCreated event: OrderID=%s, CustomerID=%s",
				event.OrderID, event.CustomerID)

			// Process notification
			if err := sendNotification(&event); err != nil {
				log.Printf("Error sending notification: %v", err)
				if nackErr := msg.Nack(false, true); nackErr != nil {
					log.Printf("Error nacking message: %v", nackErr)
				}
				continue
			}

			// Acknowledge successful processing
			if ackErr := msg.Ack(false); ackErr != nil {
				log.Printf("Error acknowledging message: %v", ackErr)
			} else {
				log.Printf("Successfully sent notification for order: %s", event.OrderID)
			}
		}
	}
}

// sendNotification simulates sending a notification (email, SMS, etc.)
func sendNotification(event *OrderCreatedEvent) error {
	// Simulate notification delay
	time.Sleep(500 * time.Millisecond)

	// In a real system, this would integrate with email service, SMS gateway, etc.
	log.Printf("📧 [NOTIFICATION] Order %s created for customer %s", event.OrderID, event.CustomerID)
	log.Printf("   Product: %s, Quantity: %d, Total: $%.2f",
		event.ProductID, event.Quantity, event.TotalAmount)

	// Simulate occasional failures for demonstration
	// In production, this would be actual failure from external service
	// Uncomment to test retry logic:
	// if rand.Float32() < 0.1 {
	// 	return fmt.Errorf("simulated notification service failure")
	// }

	return nil
}

// connectRabbitMQ establishes connection to RabbitMQ with retry
func connectRabbitMQ(url string) (*amqp.Connection, error) {
	var conn *amqp.Connection
	var err error

	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		conn, err = amqp.Dial(url)
		if err == nil {
			log.Println("Connected to RabbitMQ successfully")
			return conn, nil
		}

		log.Printf("Failed to connect to RabbitMQ (attempt %d/%d): %v", i+1, maxRetries, err)
		time.Sleep(5 * time.Second)
	}

	return nil, fmt.Errorf("failed to connect after %d attempts: %w", maxRetries, err)
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// healthCheck handles the health check endpoint
func healthCheck(w http.ResponseWriter, _ *http.Request, conn *amqp.Connection) {
	response := map[string]interface{}{
		"status":  "healthy",
		"service": "notification-worker",
	}

	checks := map[string]string{
		"mq": "healthy",
	}

	overallHealthy := true

	// Check RabbitMQ connection
	if conn == nil || conn.IsClosed() {
		checks["mq"] = "unhealthy: connection closed"
		overallHealthy = false
	}

	response["checks"] = checks
	if !overallHealthy {
		response["status"] = "degraded"
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Error encoding health check response: %v", err)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding health check response: %v", err)
	}
}
