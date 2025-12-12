package config

import (
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/wb-go/wbf/retry"
)

type Config struct {
	DB struct {
		Host            string        `env:"POSTGRES_HOST" validate:"required"`
		Port            int           `env:"POSTGRES_PORT" validate:"required"`
		User            string        `env:"POSTGRES_USER" validate:"required"`
		Pass            string        `env:"POSTGRES_PASSWORD" validate:"required"`
		DBName          string        `env:"POSTGRES_DB" validate:"required"`
		MaxOpenConns    int           `env:"DB_MAX_OPEN_CONNS"`
		MaxIdleConns    int           `env:"DB_MAX_IDLE_CONNS"`
		ConnMaxLifetime time.Duration `env:"DB_CONN_MAX_LIFETIME"`
	}

	Server struct {
		Addr            string        `env:"SERVER_PORT" validate:"required"`
		ReadTimeout     time.Duration `env:"SERVER_READ_TIMEOUT" validate:"required"`
		WriteTimeout    time.Duration `env:"SERVER_WRITE_TIMEOUT" validate:"required"`
		IdleTimeout     time.Duration `env:"SERVER_IDLE_TIMEOUT" validate:"required"`
		ShutdownTimeout time.Duration `env:"SERVER_SHUTDOWN_TIMEOUT" validate:"required"`
	}

	Retries struct {
		Attempts int     `env:"RETRIES_ATTEMPTS" validate:"required"`
		DelayMs  int     `env:"RETRIES_DELAY_MS" validate:"required"`
		Backoff  float64 `env:"RETRIES_BACKOFF" validate:"required"`
	}

	SchedulerConfig struct {
		CleanupInterval time.Duration `env:"SCHEDULER_CLEANUP_INTERVAL" validate:"required"`
		BookingTTL      time.Duration `env:"SCHEDULER_BOOKING_TTL" validate:"required"`
	}

	EmailConfig struct {
		SMTPHost     string `env:"SMTP_HOST" validate:"required"`
		SMTPPort     int    `env:"SMTP_PORT" validate:"required"`
		SMTPUser     string `env:"SMTP_USER" validate:"required"`
		SMTPPassword string `env:"SMTP_PASSWORD" validate:"required"`
		FromEmail    string `env:"FROM_EMAIL" validate:"required"`
	}

	TelegramConfig struct {
		BotToken string `env:"TELEGRAM_BOT_TOKEN" validate:"required"`
		ChatID   string `env:"TELEGRAM_CHAT_ID" validate:"required"`
	}
}

func MustLoad() (*Config, error) {
	var cfg Config

	err := cleanenv.ReadEnv(&cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to read environment variables: %w", err)
	}

	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

func (c *Config) DBDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		c.DB.User, c.DB.Pass, c.DB.Host, c.DB.Port, c.DB.DBName)
}

func (c *Config) DefaultRetryStrategy() retry.Strategy {
	return retry.Strategy{
		Attempts: c.Retries.Attempts,
		Delay:    time.Duration(c.Retries.DelayMs) * time.Millisecond,
		Backoff:  c.Retries.Backoff,
	}
}
