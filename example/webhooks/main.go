// Basic webhook handling with the SumUp SDK.
//
// This example shows how to:
// 1. Accept webhook requests over HTTP
// 2. Verify the webhook signature with your webhook secret
// 3. Parse the payload into a typed SumUp webhook event
//
// Prerequisites:
// - `SUMUP_WEBHOOK_SECRET` environment variable with your SumUp webhook signing secret
//
// To run:
//
//	export SUMUP_WEBHOOK_SECRET="whsec_test"
//	go run main.go
//
// Then send POST requests to http://localhost:8080/webhooks.
package main

import (
	"errors"
	"log"
	"net/http"

	"github.com/sumup/sumup-go"
)

func main() {
	client := sumup.NewClient()
	handler := client.WebhookHandler()

	http.HandleFunc("/webhooks", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		event, err := handler.ParseRequest(r)
		if err != nil {
			status := http.StatusUnauthorized
			if errors.Is(err, sumup.ErrWebhookSignatureExpired) {
				status = http.StatusBadRequest
			}
			if errors.Is(err, sumup.ErrWebhookSignatureInvalid) ||
				errors.Is(err, sumup.ErrWebhookTimestampInvalid) {
				http.Error(w, "Invalid webhook signature", status)
				return
			}
			http.Error(w, "Invalid webhook payload", http.StatusBadRequest)
			return
		}

		switch evt := event.(type) {
		case *sumup.CheckoutProcessedWebhookEvent:
			log.Printf("[INFO] checkout processed: event_id=%s checkout_id=%s", evt.ID, evt.Object.ID)
		case *sumup.CheckoutFailedWebhookEvent:
			log.Printf("[INFO] checkout failed: event_id=%s checkout_id=%s", evt.ID, evt.Object.ID)
		case *sumup.MemberCreatedWebhookEvent:
			log.Printf("[INFO] member created: event_id=%s member_id=%s", evt.ID, evt.Object.ID)
		case *sumup.UnknownEventNotification:
			log.Printf("[INFO] webhook received: event_id=%s type=%s object_id=%s", evt.ID, evt.Type, evt.Object.ID)
		default:
			log.Printf("[INFO] webhook received: type=%T", event)
		}

		w.WriteHeader(http.StatusOK)
	})

	log.Println("Server is running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
