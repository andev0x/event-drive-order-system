package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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

	// Connect to RabbitMQ
	conn, err := connectRabbitMQ(rabbitMQURL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	// Create channel
	channel, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open channel: %v", err)
	}
	defer channel.Close()

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
		log.Fatalf("Failed to declare exchange: %v", err)
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
		log.Fatalf("Failed to declare queue: %v", err)
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
		log.Fatalf("Failed to bind queue: %v", err)
	}

	log.Printf("Notification worker connected to queue '%s'", queueName)

	// Set QoS
	err = channel.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		log.Fatalf("Failed to set QoS: %v", err)
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
		log.Fatalf("Failed to register consumer: %v", err)
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
				msg.Nack(false, false) // Don't requeue invalid messages
				continue
			}

			log.Printf("Received OrderCreated event: OrderID=%s, CustomerID=%s",
				event.OrderID, event.CustomerID)

			// Process notification
			if err := sendNotification(&event); err != nil {
				log.Printf("Error sending notification: %v", err)
				msg.Nack(false, true) // Requeue for retry
				continue
			}

			// Acknowledge successful processing
			msg.Ack(false)
			log.Printf("Successfully sent notification for order: %s", event.OrderID)
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
