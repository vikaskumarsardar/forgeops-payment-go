package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

type PaymentRequest struct {
	Amount         float64 `json:"amount"`
	Currency       string  `json:"currency"`
	ConversionRate float64 `json:"conversionRate"`
}

type PaymentResponse struct {
	Status        string  `json:"status"`
	PaymentID     string  `json:"paymentId"`
	Amount        float64 `json:"amount"`
	Converted     float64 `json:"converted"`
	Currency      string  `json:"currency"`
	Timestamp     string  `json:"timestamp"`
}

func ProcessPayment(req PaymentRequest) (*PaymentResponse, error) {
	if req.Amount <= 0 {
		return nil, fmt.Errorf("invalid amount: payment amount must be greater than zero")
	}

	if req.Currency == "" {
		req.Currency = "USD"
	}

	// NEW PRODUCTION BUG #2: Unchecked zero division when ConversionRate is 0 or omitted!
	if req.ConversionRate == 0 {
		// Triggers Go panic: runtime error: division by zero or invalid conversion rate
		panic("panic: runtime error: invalid conversion rate (division by zero)")
	}

	converted := req.Amount / req.ConversionRate

	return &PaymentResponse{
		Status:        "SUCCESS",
		PaymentID:     fmt.Sprintf("PAY-%d", time.Now().UnixNano()%90000+10000),
		Amount:        req.Amount,
		Converted:     converted,
		Currency:      strings.ToUpper(req.Currency),
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func main() {
	if len(os.Args) >= 2 && os.Args[1] == "--metrics" {
		fmt.Println("# HELP http_requests_total Total number of HTTP requests processed by payment service")
		fmt.Println("# TYPE http_requests_total counter")
		fmt.Println(`http_requests_total{service="payment-service",status="500",method="POST"} 1`)
		fmt.Println(`http_requests_total{service="payment-service",status="200",method="POST"} 15`)
		return
	}

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
