package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

type PaymentRequest struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

type PaymentResponse struct {
	Status    string  `json:"status"`
	PaymentID string  `json:"paymentId"`
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
	Timestamp string  `json:"timestamp"`
}

func ProcessPayment(req PaymentRequest) (*PaymentResponse, error) {
	if req.Amount <= 0 {
		return nil, fmt.Errorf("invalid amount: payment amount must be greater than zero")
	}

	// INTENTIONAL PRODUCTION BUG: Panics when currency string is empty or null!
	if req.Currency == "" {
		panic("panic: currency cannot be empty")
	}

	return &PaymentResponse{
		Status:    "SUCCESS",
		PaymentID: fmt.Sprintf("PAY-%d", time.Now().UnixNano()%90000+10000),
		Amount:    req.Amount,
		Currency:  strings.ToUpper(req.Currency),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println(`{"error": "Usage: go run main.go '<json_payload>'"}`)
		os.Exit(1)
	}

	var req PaymentRequest
	err := json.Unmarshal([]byte(os.Args[1]), &req)
	if err != nil {
		fmt.Printf(`{"error": "JSON parse failed: %s"}`+"\n", err.Error())
		os.Exit(1)
	}

	res, err := ProcessPayment(req)
	if err != nil {
		fmt.Printf(`{"error": "%s"}`+"\n", err.Error())
		os.Exit(1)
	}

	out, _ := json.Marshal(res)
	fmt.Println(string(out))
}
