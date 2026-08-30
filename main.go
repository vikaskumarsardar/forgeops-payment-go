package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

type PaymentRequest struct {
	Amount         float64 `json:"amount"`
	Currency       string  `json:"currency"`
	ConversionRate float64 `json:"conversionRate"`
}

type PaymentResponse struct {
	Status    string  `json:"status"`
	PaymentID string  `json:"paymentId"`
	Amount    float64 `json:"amount"`
	Converted float64 `json:"converted"`
	Currency  string  `json:"currency"`
	Timestamp string  `json:"timestamp"`
}

var (
	total200Count uint64 = 15
	total500Count uint64 = 0
)

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
		Status:    "SUCCESS",
		PaymentID: fmt.Sprintf("PAY-%d", time.Now().UnixNano()%90000+10000),
		Amount:    req.Amount,
		Converted: converted,
		Currency:  strings.ToUpper(req.Currency),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	fmt.Fprintf(w, "# HELP http_requests_total Total number of HTTP requests processed by payment service\n")
	fmt.Fprintf(w, "# TYPE http_requests_total counter\n")
	fmt.Fprintf(w, "http_requests_total{service=\"payment-service\",status=\"200\",method=\"POST\",path=\"/api/v1/payment\"} %d\n", atomic.LoadUint64(&total200Count))
	fmt.Fprintf(w, "http_requests_total{service=\"payment-service\",status=\"500\",method=\"POST\",path=\"/api/v1/payment\"} %d\n", atomic.LoadUint64(&total500Count))
}

func paymentApiHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req PaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		atomic.AddUint64(&total500Count, 1)
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	defer func() {
		if rec := recover(); rec != nil {
			atomic.AddUint64(&total500Count, 1)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `{"error": "%v"}`, rec)
		}
	}()

	res, err := ProcessPayment(req)
	if err != nil {
		atomic.AddUint64(&total500Count, 1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	atomic.AddUint64(&total200Count, 1)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func main() {
	if len(os.Args) >= 2 && os.Args[1] == "--metrics" {
		fmt.Println("# HELP http_requests_total Total number of HTTP requests processed by payment service")
		fmt.Println("# TYPE http_requests_total counter")
		fmt.Printf("http_requests_total{service=\"payment-service\",status=\"500\",method=\"POST\"} %d\n", atomic.LoadUint64(&total500Count))
		fmt.Printf("http_requests_total{service=\"payment-service\",status=\"200\",method=\"POST\"} %d\n", atomic.LoadUint64(&total200Count))
		return
	}

	// CLI invocation mode for sandbox child processes
	if len(os.Args) >= 2 && strings.HasPrefix(os.Args[1], "{") {
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
		return
	}

	// Standalone Go Microservice Server on Port 5000
	portStr := os.Getenv("PAYMENT_SERVICE_PORT")
	if portStr == "" {
		portStr = "5000"
	}
	port, _ := strconv.Atoi(portStr)

	http.HandleFunc("/metrics", metricsHandler)
	http.HandleFunc("/api/v1/payment", paymentApiHandler)

	fmt.Printf("🚀 Payment Go Microservice Server listening on http://localhost:%d (Metrics on /metrics)\n", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		fmt.Printf("Payment service server error: %v\n", err)
	}
}
