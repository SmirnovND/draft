package config

import (
	"github.com/SmirnovND/gobase/internal/interfaces"
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

type Config struct {
	Db  `yaml:"db"`
	App `yaml:"app"`
}

type Db struct {
	Dsn string `yaml:"dsn"`
}

type App struct {
	RunAddr string `yaml:"run_addr"`
}

func (c *Config) GetDBDsn() string {
	return c.Db.Dsn
}

func (c *Config) GetRunAddr() string {
	return c.App.RunAddr
}

func NewConfig() interfaces.ConfigServer {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
		}
	}()

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
