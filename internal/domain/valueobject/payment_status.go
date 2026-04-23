// Package valueobject holds immutable value objects for the payment domain.
// PaymentStatus encodes the state machine — all valid transitions live here, in one place.
package valueobject

import "errors"

// PaymentStatus represents the lifecycle state of a payment.
// It's a string type so it maps cleanly to the DB column without extra marshaling.
type PaymentStatus string

const (
	// PaymentStatusPending is the initial status assigned on payment creation.
	PaymentStatusPending PaymentStatus = "pending"
	// PaymentStatusProcessing means the provider accepted the charge and is working on it.
	PaymentStatusProcessing PaymentStatus = "processing"
	// PaymentStatusCompleted means money was successfully captured. Safe to fulfill the order.
	PaymentStatusCompleted PaymentStatus = "completed"
	// PaymentStatusFailed is terminal — something went wrong and this payment can't be retried.
	// The caller should create a new payment with a different external_id.
	PaymentStatusFailed PaymentStatus = "failed"
	// PaymentStatusRefunded is terminal — the charge was reversed.
	PaymentStatusRefunded PaymentStatus = "refunded"
)

// validTransitions is the state machine table. Each key maps to the statuses it can move to.
// Failed and refunded have empty slices — they are exit points with no way out.
var validTransitions = map[PaymentStatus][]PaymentStatus{
	PaymentStatusPending:    {PaymentStatusProcessing, PaymentStatusFailed},
	PaymentStatusProcessing: {PaymentStatusCompleted, PaymentStatusFailed},
	PaymentStatusCompleted:  {PaymentStatusRefunded},
	PaymentStatusFailed:     {},
	PaymentStatusRefunded:   {},
}

// Validate checks if the status value is a known one.
// Useful when hydrating from the database — catches corrupted or unknown status strings early.
func (s PaymentStatus) Validate() error {
	if _, ok := validTransitions[s]; !ok {
		return errors.New("invalid payment status: " + string(s))
	}
	return nil
}

// CanTransitionTo checks whether moving from the current status to next is allowed.
// This is the enforcement point of the state machine — called by Payment.Transition().
func (s PaymentStatus) CanTransitionTo(next PaymentStatus) error {
	allowed, ok := validTransitions[s]
	if !ok {
		return errors.New("invalid current status: " + string(s))
	}
	for _, a := range allowed {
		if a == next {
			return nil
		}
	}
	return errors.New("invalid transition from " + string(s) + " to " + string(next))
}

func (s PaymentStatus) String() string { return string(s) }
