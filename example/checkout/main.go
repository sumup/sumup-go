package main

import (
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"

	gonanoid "github.com/matoous/go-nanoid/v2"

	"github.com/sumup/sumup-go"
	"github.com/sumup/sumup-go/checkouts"
)

var (
	//go:embed templates
	templatesFs embed.FS
	templates   *template.Template
)

func init() {
	templates = template.Must(template.ParseFS(templatesFs, "templates/*.html"))
}

func main() {
	merchantCode := os.Getenv("SUMUP_MERCHANT_CODE")
	client := sumup.NewClient()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		templates.ExecuteTemplate(w, "index.html", nil)
	})
	http.HandleFunc("/create-checkout", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		rawAmount := r.FormValue("amount")
		amount, err := strconv.ParseFloat(rawAmount, 64)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid amount %q: %v.", rawAmount, err), http.StatusBadRequest)
			return
		}

		checkoutReference := gonanoid.Must()
		description := "Test Payment"
		checkout, err := client.Checkouts.Create(r.Context(), checkouts.CreateCheckoutBody{
			Amount:            amount,
			CheckoutReference: checkoutReference,
			Currency:          "EUR",
			MerchantCode:      merchantCode,
			Description:       &description,
		})
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to create checkout: %v.", err), http.StatusInternalServerError)
			return
		}

		data := struct {
			CheckoutID string
		}{
			CheckoutID: *checkout.Id,
		}

		templates.ExecuteTemplate(w, "checkout.html", data)
	})
	http.HandleFunc("/payment-result", func(w http.ResponseWriter, r *http.Request) {
		status := r.URL.Query().Get("status")
		message := r.URL.Query().Get("message")

		templates.ExecuteTemplate(w, "result.html", struct {
			Status  string
			Message string
		}{
			Status:  status,
			Message: message,
		})
	})

	fmt.Println("Server started at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
