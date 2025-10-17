package config

import (
	"github.com/SmirnovND/gobase/internal/interfaces"
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

type Config struct {
	Db       `yaml:"db"`
	App      `yaml:"app"`
	RabbitMQ `yaml:"rabbitmq"`
}

type Db struct {
	Dsn          string `yaml:"dsn"`
	MaxOpenConns int    `yaml:"max_open_conns"`
	MaxIdleConns int    `yaml:"max_idle_conns"`
}

type App struct {
	RunAddr string `yaml:"run_addr"`
}

type RabbitMQ struct {
	URL string `yaml:"url"`
}

func (c *Config) GetDBDsn() string {
	return c.Db.Dsn
}

func (c *Config) GetDBMaxOpenConns() int {
	return c.Db.MaxOpenConns
}

func (c *Config) GetDBMaxIdleConns() int {
	return c.Db.MaxIdleConns
}

func (c *Config) GetRunAddr() string {
	return c.App.RunAddr
}

func (c *Config) GetRabbitMQURL() string {
	return c.RabbitMQ.URL
}

func NewConfig() interfaces.ConfigServer {
	if len(os.Args) < 2 {
		log.Fatal("Config file path not provided. Usage: ./server <config.yaml>")
	}

	cPath := os.Args[1]
	cf := &Config{}
	cf.LoadConfig(cPath)

	return cf
}

func (c *Config) LoadConfig(patch string) {
	file, err := os.Open(patch)
	if err != nil {
		log.Fatal("ReadConfigFile: ", err)
	}

	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(&c)
	if err != nil {
		log.Fatal("DecodeConfigFile: ", err)
	}
}
