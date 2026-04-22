package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"

	"github.com/ademarthiago/payment-gateway/internal/adapter/event"
	"github.com/ademarthiago/payment-gateway/internal/infrastructure/migration"
	adapterhttp "github.com/ademarthiago/payment-gateway/internal/adapter/http"
	"github.com/ademarthiago/payment-gateway/internal/adapter/http/handler"
	"github.com/ademarthiago/payment-gateway/internal/adapter/postgres"
	adapterredis "github.com/ademarthiago/payment-gateway/internal/adapter/redis"
	"github.com/ademarthiago/payment-gateway/internal/domain/port"
	"github.com/ademarthiago/payment-gateway/internal/usecase"
	"github.com/ademarthiago/payment-gateway/pkg/logger"
)

var version = "dev"

func main() {
	// Load .env (non-fatal in production)
	_ = godotenv.Load()

	// Setup logger
	logger.Setup(os.Getenv("APP_ENV"))
	log.Info().Str("version", version).Msg("starting payment-gateway")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// --- Infrastructure ---

	// PostgreSQL
	pgPool, err := postgres.NewPool(ctx, postgres.ConfigFromEnv())
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to postgres")
	}
	defer pgPool.Close()
	log.Info().Msg("postgres connected")

	// Run migrations
	migrationDSN := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s&search_path=payment,public",
		os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"), os.Getenv("DB_SSL_MODE"),
	)
	if err := migration.Run(migrationDSN); err != nil {
		log.Fatal().Err(err).Msg("failed to run migrations")
	}
	log.Info().Msg("migrations applied")

	// Redis
	redisClient, err := adapterredis.NewClient(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to redis")
	}
	defer redisClient.Close()
	log.Info().Msg("redis connected")

	// --- Adapters ---
	paymentRepo := postgres.NewPaymentRepository(pgPool)
	outboxRepo := postgres.NewOutboxRepository(pgPool)
	idempotencyStore := adapterredis.NewIdempotencyStore(redisClient)

	// Event channel (buffered)
	eventCh := make(chan port.Event, 256)
	publisher := event.NewChannelPublisher(eventCh)
	dispatcher := event.NewDispatcher(eventCh)

	// Outbox worker interval
	interval := 5 * time.Second
	if s := os.Getenv("OUTBOX_WORKER_INTERVAL_SECONDS"); s != "" {
		if n, err := strconv.Atoi(s); err == nil {
			interval = time.Duration(n) * time.Second
		}
	}
	outboxWorker := event.NewOutboxWorker(outboxRepo, publisher, interval)

	// --- Use Cases ---
	createPaymentUC := usecase.NewCreatePaymentUseCase(paymentRepo, outboxRepo, idempotencyStore, publisher)
	getPaymentUC := usecase.NewGetPaymentUseCase(paymentRepo)
	processRefundUC := usecase.NewProcessRefundUseCase(paymentRepo, outboxRepo, publisher)

	// --- HTTP ---
	healthHandler := handler.NewHealthHandler(pgPool, adapterredis.NewHealthChecker(redisClient))
	paymentHandler := handler.NewPaymentHandler(createPaymentUC, getPaymentUC, processRefundUC)
	router := adapterhttp.NewRouter(paymentHandler, healthHandler)

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8088"
	}

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// --- Background workers ---
	go dispatcher.Start(ctx)
	go outboxWorker.Start(ctx)

	// --- Start server ---
	go func() {
		log.Info().Str("port", port).Msg("http server started")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server error")
		}
	}()

	// --- Graceful shutdown ---
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down gracefully...")
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("forced shutdown")
	}

	log.Info().Msg("server stopped")
}
