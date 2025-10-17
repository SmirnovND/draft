package adapter

import (
	"github.com/SmirnovND/toolbox/pkg/rabbitmq"
	"github.com/jmoiron/sqlx"
)

// SQLXDBCloser адаптер для *sqlx.DB
type SQLXDBCloser struct {
	db *sqlx.DB
}

func NewSQLXDBCloser(db *sqlx.DB) *SQLXDBCloser {
	return &SQLXDBCloser{db: db}
}

func (c *SQLXDBCloser) Close() error {
	return c.db.Close()
}

// RabbitMQConnectionCloser адаптер для RabbitMQConnection
type RabbitMQConnectionCloser struct {
	conn *rabbitmq.RabbitMQConnection
}

func NewRabbitMQConnectionCloser(conn *rabbitmq.RabbitMQConnection) *RabbitMQConnectionCloser {
	return &RabbitMQConnectionCloser{conn: conn}
}

func (c *RabbitMQConnectionCloser) Close() error {
	c.conn.Close()
	return nil
}

// RabbitMQProducerCloser адаптер для RabbitMQProducer
type RabbitMQProducerCloser struct {
	producer *rabbitmq.RabbitMQProducer
}

func NewRabbitMQProducerCloser(producer *rabbitmq.RabbitMQProducer) *RabbitMQProducerCloser {
	return &RabbitMQProducerCloser{producer: producer}
}

func (c *RabbitMQProducerCloser) Close() error {
	c.producer.Close()
	return nil
}

// RabbitMQConsumerCloser адаптер для RabbitMQConsumer
type RabbitMQConsumerCloser struct {
	consumer *rabbitmq.RabbitMQConsumer
}

func NewRabbitMQConsumerCloser(consumer *rabbitmq.RabbitMQConsumer) *RabbitMQConsumerCloser {
	return &RabbitMQConsumerCloser{consumer: consumer}
}

func (c *RabbitMQConsumerCloser) Close() error {
	c.consumer.Close()
	return nil
}
