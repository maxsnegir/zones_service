package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env     string        `yaml:"env" env-default:"local"`
	Storage StorageConfig `yaml:"storage" env-required:"true"`
	Server  ServerConfig  `yaml:"server"`
}

type StorageConfig struct {
	DSN string `yaml:"dsn"`
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

func MustLoad() *Config {
	configPath := fetchConfigPath()

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic("config file does not exist: " + configPath)
	}
	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic(fmt.Sprintf("failed to read config: %s", err.Error()))
	}

	return &cfg

}

func fetchConfigPath() string {
	var configPath string

	flag.StringVar(&configPath, "config", "", "path to config file")
	flag.Parse()

	if configPath == "" {
		configPath = os.Getenv("CONFIG_PATH")
	}

	return configPath
}
