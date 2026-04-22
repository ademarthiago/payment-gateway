package valueobject

import "errors"

// PaymentStatus represents the lifecycle of a payment
type PaymentStatus string

const (
	PaymentStatusPending    PaymentStatus = "pending"
	PaymentStatusProcessing PaymentStatus = "processing"
	PaymentStatusCompleted  PaymentStatus = "completed"
	PaymentStatusFailed     PaymentStatus = "failed"
	PaymentStatusRefunded   PaymentStatus = "refunded"
)

var validTransitions = map[PaymentStatus][]PaymentStatus{
	PaymentStatusPending:    {PaymentStatusProcessing, PaymentStatusFailed},
	PaymentStatusProcessing: {PaymentStatusCompleted, PaymentStatusFailed},
	PaymentStatusCompleted:  {PaymentStatusRefunded},
	PaymentStatusFailed:     {},
	PaymentStatusRefunded:   {},
}

func (s PaymentStatus) Validate() error {
	if _, ok := validTransitions[s]; !ok {
		return errors.New("invalid payment status: " + string(s))
	}
	return nil
}

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
