package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/andev0x/order-service/internal/model"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	exchangeName = "orders"
	exchangeType = "topic"
	routingKey   = "order.created"
)

// EventPublisher interface for publishing events
type EventPublisher interface {
	PublishOrderCreated(ctx context.Context, event *model.OrderCreatedEvent) error
	Close() error
}

// RabbitMQPublisher implements EventPublisher using RabbitMQ
type RabbitMQPublisher struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

// NewRabbitMQPublisher creates a new RabbitMQ publisher
func NewRabbitMQPublisher(url string) (*RabbitMQPublisher, error) {
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

	log.Printf("RabbitMQ publisher connected and exchange '%s' declared", exchangeName)

	return &RabbitMQPublisher{
		conn:    conn,
		channel: channel,
	}, nil
}

// PublishOrderCreated publishes an order created event
func (p *RabbitMQPublisher) PublishOrderCreated(ctx context.Context, event *model.OrderCreatedEvent) error {
	event.EventType = "OrderCreated"

	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	err = p.channel.PublishWithContext(
		ctx,
		exchangeName, // exchange
		routingKey,   // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	log.Printf("Published OrderCreated event for order: %s", event.OrderID)
	return nil
}

// Close closes the RabbitMQ connection
func (p *RabbitMQPublisher) Close() error {
	if p.channel != nil {
		if err := p.channel.Close(); err != nil {
			return err
		}
	}
	if p.conn != nil {
		if err := p.conn.Close(); err != nil {
			return err
		}
	}
	return nil
}

// HealthCheck checks if the RabbitMQ connection is alive
func (p *RabbitMQPublisher) HealthCheck() error {
	if p.conn == nil {
		return fmt.Errorf("connection is nil")
	}
	if p.conn.IsClosed() {
		return fmt.Errorf("connection is closed")
	}
	if p.channel == nil {
		return fmt.Errorf("channel is nil")
	}
	return nil
}
