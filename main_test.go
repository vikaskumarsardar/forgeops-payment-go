package main

import (
	"testing"
)

func TestValidPayment(t *testing.T) {
	req := PaymentRequest{
		Amount:   150.00,
		Currency: "USD",
	}

	res, err := ProcessPayment(req)
	if err != nil {
		t.Fatalf("Expected valid payment, got error: %v", err)
	}

	if res.Status != "SUCCESS" {
		t.Errorf("Expected status SUCCESS, got %s", res.Status)
	}
}

func TestMissingCurrencyPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic on empty currency, but code did not panic")
		}
	}()

	req := PaymentRequest{
		Amount:   50.00,
		Currency: "",
	}

	_, _ = ProcessPayment(req)
}
