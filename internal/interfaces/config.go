package interfaces

type ConfigServer interface {
	GetDBDsn() string
	GetRunAddr() string
	GetRabbitMQURL() string
}
