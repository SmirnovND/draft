package interfaces

type ConfigServer interface {
	GetDBDsn() string
	GetDBMaxOpenConns() int
	GetDBMaxIdleConns() int
	GetRunAddr() string
	GetRabbitMQURL() string
}
