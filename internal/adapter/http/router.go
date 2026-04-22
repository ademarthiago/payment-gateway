package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/ademarthiago/payment-gateway/internal/adapter/http/handler"
	"github.com/ademarthiago/payment-gateway/internal/adapter/http/middleware"
)

func NewRouter(
	paymentHandler *handler.PaymentHandler,
	healthHandler *handler.HealthHandler,
) http.Handler {
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(middleware.Recovery)
	r.Use(middleware.Logging)
	r.Use(chimiddleware.Compress(5))

	// Health check
	r.Get("/health", healthHandler.Handle)
	r.Head("/health", healthHandler.Handle)

	// API v1
	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/payments", func(r chi.Router) {
			r.Post("/", paymentHandler.CreatePayment)
			r.Get("/{id}", paymentHandler.GetPayment)
			r.Post("/{id}/refund", paymentHandler.ProcessRefund)
		})
	})

	return r
}
