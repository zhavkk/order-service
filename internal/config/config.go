package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env      string         `yaml:"env" env:"ENV" env-default:"local"`
	HTTP     HTTPConfig     `yaml:"http"`
	Postgres PostgresConfig `yaml:"postgres"`
	Redis    RedisConfig    `yaml:"redis"`
	Kafka    KafkaConfig    `yaml:"kafka"`
}

type HTTPConfig struct {
	Port string `yaml:"port" env:"HTTP_PORT" env-default:":8080"`
}

type PostgresConfig struct {
	Host     string `env:"POSTGRES_HOST" envDefault:"localhost"`
	Port     string `env:"POSTGRES_PORT" envDefault:"5432"`
	Username string `env:"POSTGRES_USER" envDefault:"postgres"`
	Password string `env:"POSTGRES_PASSWORD" envDefault:"postgres"`
	Database string `env:"POSTGRES_DB" envDefault:"L0_test_service"`
	SSLMode  string `env:"POSTGRES_SSLMODE" envDefault:"disable"`
}

type RedisConfig struct {
	Host string        `env:"REDIS_HOST" envDefault:"localhost"`
	Port string        `env:"REDIS_PORT" envDefault:"6379"`
	TTL  time.Duration `yaml:"ttl" env:"REDIS_TTL" env-default:"5m"`
}

type KafkaConfig struct {
	Version            string        `yaml:"version" env:"KAFKA_VERSION" env-default:"2.8.0"`
	AutoCommitInterval time.Duration `yaml:"auto_commit_interval" env:"KAFKA_AUTO_COMMIT_INTERVAL" env-default:"1s"`
}

func (r RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%s", r.Host, r.Port)
}

func (c PostgresConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.Username, c.Password, c.Host, c.Port, c.Database, c.SSLMode,
	)
}

func MustLoad(configPath string) *Config {
	if configPath == "" {
		panic("Config path is not set")
	}
	if _, err := os.Stat(configPath); err != nil {
		log.Fatalf("config file does not exist: %s", configPath)
	}
	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("can't read config %s", err.Error())
	}
	return &cfg
}
