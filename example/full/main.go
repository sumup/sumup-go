// This example demonstrates a complete checkout flow with the SumUp payment widget.
//
// Basic steps:
// 1. Create a checkout for the desired amount using the SDK
// 2. Initiate the payment widget with the checkout ID
//
// Prerequisites:
// - SUMUP_API_KEY environment variable with your SumUp API key
// - SUMUP_MERCHANT_CODE environment variable with your merchant code
//
// To run:
//
//	export SUMUP_API_KEY="your-api-key"
//	export SUMUP_MERCHANT_CODE="your-merchant-code"
//	go run main.go
//
// Then open http://localhost:8080 in your browser.
package main

import (
	"cmp"
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"math/rand"
	"net/http"
	"os"

	"github.com/sumup/sumup-go"
)

//go:embed templates/*.html
var templateFS embed.FS

//go:embed static/*
var staticFS embed.FS

func generateRandomCheckoutID() string {
	return fmt.Sprintf("CHK-%d", rand.Int())
}

func main() {
	apiKey := os.Getenv("SUMUP_API_KEY")
	if apiKey == "" {
		slog.Error("SUMUP_API_KEY environment variable is required")
		os.Exit(1)
	}

	merchantCode := os.Getenv("SUMUP_MERCHANT_CODE")
	if merchantCode == "" {
		slog.Error("SUMUP_MERCHANT_CODE environment variable is required")
		os.Exit(1)
	}

	ctx := context.Background()
	client := sumup.NewClient()

	merchant, err := client.Merchants.Get(ctx, merchantCode, sumup.MerchantsGetParams{})
	if err != nil {
		slog.Error("Failed to load merchant information", "error", err)
		os.Exit(1)
	}

	slog.Info("Merchant information loaded",
		"name", *cmp.Or(merchant.BusinessProfile.Name, merchant.Company.Name),
		"code", merchant.MerchantCode,
		"currency", merchant.DefaultCurrency)

	tmpl, err := template.ParseFS(templateFS, "templates/*.html")
	if err != nil {
		slog.Error("Failed to parse templates", "error", err)
		os.Exit(1)
	}

	staticContent, err := fs.Sub(staticFS, "static")
	if err != nil {
		slog.Error("Failed to create static file system", "error", err)
		os.Exit(1)
	}
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticContent))))

	indexHandler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		if err := tmpl.ExecuteTemplate(w, "index.html", merchant); err != nil {
			slog.Error("Failed to execute template", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	}
	http.HandleFunc("/", indexHandler)

	// Handle checkout creation endpoint
	http.HandleFunc("/checkout", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse request body
		var req struct {
			Amount float32 `json:"amount"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			slog.Error("Failed to decode request", "error", err)
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		ctx := r.Context()

		checkoutID := generateRandomCheckoutID()

		// Create checkout using the SDK
		checkout, err := client.Checkouts.Create(ctx, sumup.CheckoutsCreateParams{
			MerchantCode:      merchant.MerchantCode,
			Amount:            req.Amount,
			Currency:          sumup.Currency(merchant.DefaultCurrency),
			CheckoutReference: checkoutID,
		})
		if eErr := new(sumup.ErrorExtended); errors.As(err, &eErr) {
			slog.Error("Failed to create checkout", "error_code", *eErr.ErrorCode, "message", *eErr.Message)
			http.Error(w, "Failed to create checkout", http.StatusInternalServerError)
			return
		} else if err != nil {
			slog.Error("Failed to create checkout", "error", err)
			http.Error(w, "Failed to create checkout", http.StatusInternalServerError)
			return
		}

		slog.Info("Checkout created", "id", *checkout.ID, "amount", req.Amount, "currency", merchant.DefaultCurrency, "reference", checkoutID)

		// Return checkout ID to the client
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"checkoutId": *checkout.ID,
		})
	})

	slog.Info("Server started at http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		slog.Error("Server failed", "error", err)
		os.Exit(1)
	}
}
