package valueobject_test

import (
	"testing"

	"github.com/ademarthiago/payment-gateway/internal/domain/valueobject"
)

func TestPaymentStatus_ValidTransitions(t *testing.T) {
	tests := []struct {
		from    valueobject.PaymentStatus
		to      valueobject.PaymentStatus
		wantErr bool
	}{
		{valueobject.PaymentStatusPending, valueobject.PaymentStatusProcessing, false},
		{valueobject.PaymentStatusPending, valueobject.PaymentStatusFailed, false},
		{valueobject.PaymentStatusProcessing, valueobject.PaymentStatusCompleted, false},
		{valueobject.PaymentStatusProcessing, valueobject.PaymentStatusFailed, false},
		{valueobject.PaymentStatusCompleted, valueobject.PaymentStatusRefunded, false},
		// Invalid transitions
		{valueobject.PaymentStatusPending, valueobject.PaymentStatusCompleted, true},
		{valueobject.PaymentStatusCompleted, valueobject.PaymentStatusPending, true},
		{valueobject.PaymentStatusFailed, valueobject.PaymentStatusCompleted, true},
		{valueobject.PaymentStatusRefunded, valueobject.PaymentStatusPending, true},
	}

	for _, tt := range tests {
		err := tt.from.CanTransitionTo(tt.to)
		if tt.wantErr && err == nil {
			t.Errorf("expected error for transition %s -> %s", tt.from, tt.to)
		}
		if !tt.wantErr && err != nil {
			t.Errorf("unexpected error for transition %s -> %s: %v", tt.from, tt.to, err)
		}
	}
}
