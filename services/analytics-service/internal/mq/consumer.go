package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/andev0x/analytics-service/internal/model"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	exchangeName = "orders"
	exchangeType = "topic"
	queueName    = "analytics.orders"
	routingKey   = "order.created"
)

// EventConsumer interface for consuming events
type EventConsumer interface {
	StartConsuming(ctx context.Context, handler func(*model.OrderCreatedEvent) error) error
	Close() error
}

// RabbitMQConsumer implements EventConsumer using RabbitMQ
type RabbitMQConsumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

// NewRabbitMQConsumer creates a new RabbitMQ consumer
func NewRabbitMQConsumer(url string) (*RabbitMQConsumer, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Declare exchange
	err = channel.ExchangeDeclare(
		exchangeName, // name
		exchangeType, // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Declare queue
	queue, err := channel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	// Bind queue to exchange
	err = channel.QueueBind(
		queue.Name,   // queue name
		routingKey,   // routing key
		exchangeName, // exchange
		false,
		nil,
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to bind queue: %w", err)
	}

	log.Printf("RabbitMQ consumer connected, queue '%s' bound to exchange '%s'", queueName, exchangeName)

	return &RabbitMQConsumer{
		conn:    conn,
		channel: channel,
	}, nil
}

// StartConsuming starts consuming messages from the queue
func (c *RabbitMQConsumer) StartConsuming(ctx context.Context, handler func(*model.OrderCreatedEvent) error) error {
	// Set QoS to process one message at a time
	err := c.channel.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	msgs, err := c.channel.Consume(
		queueName, // queue
		"",        // consumer
		false,     // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	log.Println("Analytics service is now consuming order events...")

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("Stopping consumer...")
				return
			case msg, ok := <-msgs:
				if !ok {
					log.Println("Message channel closed")
					return
				}

				// Parse event
				var event model.OrderCreatedEvent
				if err := json.Unmarshal(msg.Body, &event); err != nil {
					log.Printf("Error unmarshaling event: %v", err)
					msg.Nack(false, false) // Don't requeue invalid messages
					continue
				}

				log.Printf("Received OrderCreated event: OrderID=%s, CustomerID=%s, Amount=%.2f",
					event.OrderID, event.CustomerID, event.TotalAmount)

				// Process event
				if err := handler(&event); err != nil {
					log.Printf("Error processing event: %v", err)
					// Requeue the message for retry
					msg.Nack(false, true)
					continue
				}

				// Acknowledge successful processing
				msg.Ack(false)
				log.Printf("Successfully processed event for order: %s", event.OrderID)
			}
		}
	}()

	return nil
}

// Close closes the RabbitMQ connection
func (c *RabbitMQConsumer) Close() error {
	if c.channel != nil {
		if err := c.channel.Close(); err != nil {
			return err
		}
	}
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			return err
		}
	}
	return nil
}

// HealthCheck checks if the RabbitMQ connection is alive
func (c *RabbitMQConsumer) HealthCheck() error {
	if c.conn == nil {
		return fmt.Errorf("connection is nil")
	}
	if c.conn.IsClosed() {
		return fmt.Errorf("connection is closed")
	}
	if c.channel == nil {
		return fmt.Errorf("channel is nil")
	}
	return nil
}
