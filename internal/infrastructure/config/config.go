package config

import (
	"fmt"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

// Configuration represents YAML configuration of the application
type Configuration struct {
	Environment        string             `yaml:"env" env-default:"local"`
	HTTPServer         HTTPServer         `yaml:"http_server" env-required:"true"`
	PostgresConnection PostgresConnection `yaml:"postgres_connection" env-required:"true"`
	JWT                JWT                `yaml:"jwt" env-required:"true"`
}

// HTTPServer represents config of the application server
type HTTPServer struct {
	Address     string        `yaml:"address" env-required:"true"`
	Timeout     time.Duration `yaml:"timeout" env-required:"true"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-required:"true"`
}

// PostgresConnection represents config of postgres credentials
type PostgresConnection struct {
	Host     string `yaml:"host" env-required:"true" env:"POSTGRES_HOST"`
	Port     string `yaml:"port" env-required:"true" env:"POSTGRES_PORT"`
	Username string `yaml:"username" env-required:"true"`
	Password string `yaml:"password" env-required:"true" env:"POSTGRES_PASSWORD"`
	DBName   string `yaml:"db_name" env-required:"true"`
	SSLMode  string `yaml:"ssl_mode" env-required:"true"`
}

type JWT struct {
	Secret string        `yaml:"secret" env-required:"true" env:"JWT_SECRET"`
	TTL    time.Duration `yaml:"ttl" env-required:"true"`
	Issuer string        `yaml:"issuer" env-required:"true"`
}

// MustLoad loads the configuration from the file,
// which path is given in CONFIG_PATH environment variable
func MustLoad() Configuration {
	if err := godotenv.Load(".env"); err != nil {
		panic(fmt.Sprintf("Error loading .env file: %s", err))
	}

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		panic("CONFIG_PATH environment variable not set")
	}

	if _, err := os.Stat(configPath); err != nil {
		panic(fmt.Sprintf("file with path %s does not exist", configPath))
	}

	var cfg Configuration

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic(fmt.Sprintf("error reading config: %s", err))
	}

	return cfg
}
