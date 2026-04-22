package postgres

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	Host         string
	Port         string
	DBName       string
	User         string
	Password     string
	SSLMode      string
	MaxOpenConns int
	MaxIdleConns int
}

func ConfigFromEnv() Config {
	maxOpen, _ := strconv.Atoi(os.Getenv("DB_MAX_OPEN_CONNS"))
	if maxOpen == 0 {
		maxOpen = 25
	}
	maxIdle, _ := strconv.Atoi(os.Getenv("DB_MAX_IDLE_CONNS"))
	if maxIdle == 0 {
		maxIdle = 5
	}
	return Config{
		Host:         os.Getenv("DB_HOST"),
		Port:         os.Getenv("DB_PORT"),
		DBName:       os.Getenv("DB_NAME"),
		User:         os.Getenv("DB_USER"),
		Password:     os.Getenv("DB_PASSWORD"),
		SSLMode:      os.Getenv("DB_SSL_MODE"),
		MaxOpenConns: maxOpen,
		MaxIdleConns: maxIdle,
	}
}

func NewPool(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s dbname=%s user=%s password=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.DBName, cfg.User, cfg.Password, cfg.SSLMode,
	)

	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse db config: %w", err)
	}

	poolCfg.MaxConns = int32(cfg.MaxOpenConns)
	poolCfg.MinConns = int32(cfg.MaxIdleConns)
	poolCfg.HealthCheckPeriod = 1 * time.Minute
	poolCfg.ConnConfig.ConnectTimeout = 5 * time.Second

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}
