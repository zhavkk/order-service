package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Env      string         `yaml:"env" env:"ENV" env-default:"local"`
	HTTP     HTTPConfig     `yaml:"http"`
	Postgres PostgresConfig `yaml:"postgres"`
	Redis    RedisConfig    `yaml:"redis"`
	Kafka    KafkaConfig    `yaml:"kafka"`
}

type HTTPConfig struct {
	Port         int           `yaml:"port" env:"HTTP_PORT" env-default:"8080"`
	ReadTimeout  time.Duration `yaml:"read_timeout" env:"HTTP_READ_TIMEOUT" env-default:"5s"`
	WriteTimeout time.Duration `yaml:"write_timeout" env:"HTTP_WRITE_TIMEOUT" env-default:"5s"`
	IdleTimeout  time.Duration `yaml:"idle_timeout" env:"HTTP_IDLE_TIMEOUT" env-default:"60s"`
}

type PostgresConfig struct {
	Host     string        `env:"POSTGRES_HOST" envDefault:"localhost"`
	Port     string        `env:"POSTGRES_PORT" envDefault:"5432"`
	Username string        `env:"POSTGRES_USER" envDefault:"postgres"`
	Password string        `env:"POSTGRES_PASSWORD" envDefault:"postgres"`
	Database string        `env:"POSTGRES_DB" envDefault:"order_service"`
	SSLMode  string        `env:"POSTGRES_SSLMODE" envDefault:"disable"`
	Retries  int           `yaml:"retries" env:"POSTGRES_RETRY_COUNT" env-default:"3"`
	Backoff  time.Duration `yaml:"backoff" env:"POSTGRES_BACKOFF" env-default:"1s"`
}

type RedisConfig struct {
	Host string        `yaml:"host" envDefault:"localhost"`
	Port string        `yaml:"port" envDefault:"6379"`
	TTL  time.Duration `yaml:"ttl" env:"REDIS_TTL" env-default:"5m"`
	Db   int           `yaml:"db" env:"REDIS_DB" env-default:"0"`
}

type KafkaConfig struct {
	Version            string        `yaml:"version" env:"KAFKA_VERSION" env-default:"2.8.0"`
	AutoCommitInterval time.Duration `yaml:"auto_commit_interval" env:"KAFKA_AUTO_COMMIT_INTERVAL" env-default:"1s"`
	Brokers            []string      `yaml:"brokers" env:"KAFKA_BROKERS" env-default:"localhost:9092"`
	OrderTopic         string        `yaml:"order_topic" env:"KAFKA_TOPIC" env-default:"orders"`
	GroupID            string        `yaml:"group_id" env:"KAFKA_GROUP_ID" env-default:"order_service_group"`
	Retries            int           `yaml:"retries" env:"KAFKA_RETRY_COUNT" env-default:"3"`
	Backoff            time.Duration `yaml:"backoff" env:"KAFKA_BACKOFF" env-default:"1s"`
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
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found or failed to load: %v", err)
	}
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
