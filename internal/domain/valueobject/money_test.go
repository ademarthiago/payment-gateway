package valueobject_test

import (
	"testing"

	"github.com/ademarthiago/payment-gateway/internal/domain/valueobject"
)

func TestNewMoney_Success(t *testing.T) {
	m, err := valueobject.NewMoney(1000, valueobject.CurrencyBRL)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if m.Amount() != 1000 {
		t.Errorf("expected amount 1000, got %d", m.Amount())
	}
	if m.Currency() != valueobject.CurrencyBRL {
		t.Errorf("expected BRL, got %s", m.Currency())
	}
}

func TestNewMoney_InvalidAmount(t *testing.T) {
	_, err := valueobject.NewMoney(0, valueobject.CurrencyBRL)
	if err == nil {
		t.Fatal("expected error for zero amount")
	}
	_, err = valueobject.NewMoney(-100, valueobject.CurrencyBRL)
	if err == nil {
		t.Fatal("expected error for negative amount")
	}
}

func TestNewMoney_InvalidCurrency(t *testing.T) {
	_, err := valueobject.NewMoney(1000, valueobject.Currency("XYZ"))
	if err == nil {
		t.Fatal("expected error for invalid currency")
	}
}

func TestMoney_Add(t *testing.T) {
	m1, _ := valueobject.NewMoney(1000, valueobject.CurrencyBRL)
	m2, _ := valueobject.NewMoney(500, valueobject.CurrencyBRL)

	result, err := m1.Add(m2)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Amount() != 1500 {
		t.Errorf("expected 1500, got %d", result.Amount())
	}
}

func TestMoney_Add_DifferentCurrencies(t *testing.T) {
	m1, _ := valueobject.NewMoney(1000, valueobject.CurrencyBRL)
	m2, _ := valueobject.NewMoney(500, valueobject.CurrencyUSD)

	_, err := m1.Add(m2)
	if err == nil {
		t.Fatal("expected error adding different currencies")
	}
}
