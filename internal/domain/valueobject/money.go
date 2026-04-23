// Package valueobject holds immutable value objects for the payment domain.
// These types encode business rules about what values are valid — construction always validates.
package valueobject

import (
	"errors"
	"fmt"
)

// Money represents a monetary value. Amount is stored as int64 in the smallest currency unit
// (cents) to avoid float precision issues — 0.1 + 0.2 != 0.3 in floating point, and that
// kind of bug in a billing system is really bad.
type Money struct {
	amount   int64
	currency Currency
}

var (
	// ErrInvalidAmount is returned when amount is zero or negative.
	// There's no business case for charging nothing or a negative amount.
	ErrInvalidAmount = errors.New("amount must be greater than zero")
	// ErrCurrencyMissing is returned when no currency is provided.
	ErrCurrencyMissing = errors.New("currency is required")
)

// NewMoney creates a Money value, validating that amount is positive and currency is supported.
func NewMoney(amount int64, currency Currency) (Money, error) {
	if amount <= 0 {
		return Money{}, ErrInvalidAmount
	}
	if err := currency.Validate(); err != nil {
		return Money{}, err
	}
	return Money{amount: amount, currency: currency}, nil
}

func (m Money) Amount() int64      { return m.amount }
func (m Money) Currency() Currency { return m.currency }

// String formats the money as a human-readable string (e.g. "BRL 99.00").
// Not meant for financial calculations — use Amount() for arithmetic.
func (m Money) String() string {
	return fmt.Sprintf("%s %.2f", m.currency, float64(m.amount)/100)
}

// Equals returns true only if both the amount and currency match exactly.
func (m Money) Equals(other Money) bool {
	return m.amount == other.amount && m.currency == other.currency
}

// Add sums two Money values. Returns an error if currencies differ —
// silently adding BRL to USD without conversion would corrupt the amount.
func (m Money) Add(other Money) (Money, error) {
	if m.currency != other.currency {
		return Money{}, errors.New("cannot add money with different currencies")
	}
	return NewMoney(m.amount+other.amount, m.currency)
}
