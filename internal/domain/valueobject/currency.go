package valueobject

import "errors"

// Currency represents an ISO 4217 currency code
type Currency string

const (
	CurrencyBRL Currency = "BRL"
	CurrencyUSD Currency = "USD"
	CurrencyEUR Currency = "EUR"
)

var validCurrencies = map[Currency]bool{
	CurrencyBRL: true,
	CurrencyUSD: true,
	CurrencyEUR: true,
}

func (c Currency) Validate() error {
	if !validCurrencies[c] {
		return errors.New("invalid currency: " + string(c))
	}
	return nil
}

func (c Currency) String() string { return string(c) }
