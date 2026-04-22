package valueobject

import (
	"errors"
	"fmt"
)

// Money represents a monetary value stored in cents to avoid float precision issues
type Money struct {
	amount   int64
	currency Currency
}

var (
	ErrInvalidAmount   = errors.New("amount must be greater than zero")
	ErrCurrencyMissing = errors.New("currency is required")
)

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

func (m Money) String() string {
	return fmt.Sprintf("%s %.2f", m.currency, float64(m.amount)/100)
}

func (m Money) Equals(other Money) bool {
	return m.amount == other.amount && m.currency == other.currency
}

func (m Money) Add(other Money) (Money, error) {
	if m.currency != other.currency {
		return Money{}, errors.New("cannot add money with different currencies")
	}
	return NewMoney(m.amount+other.amount, m.currency)
}
