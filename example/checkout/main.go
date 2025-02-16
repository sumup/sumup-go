package main

import (
	"context"
	"log"

	"github.com/sumup/sumup-go"
	"github.com/sumup/sumup-go/checkouts"
)

func main() {
	ctx := context.Background()
	client := sumup.NewClient()

	checkout, err := client.Checkouts.Create(ctx, checkouts.CreateCheckoutBody{
		Amount:            123,
		CheckoutReference: "TX000001",
		Currency:          "EUR",
		MerchantCode:      "MK0001",
	})
	if err != nil {
		log.Printf("[ERROR] create checkout: %v", err)
		return
	}

	log.Printf("[INFO] checkout created: id=%q, amount=%v, currency=%q", *checkout.Id, *checkout.Amount, string(*checkout.Currency))

	checkoutSuccess, err := client.Checkouts.Process(ctx, *checkout.Id, checkouts.ProcessCheckoutBody{
		Card: &checkouts.Card{
			Cvv:         "123",
			ExpiryMonth: "12",
			ExpiryYear:  "2023",
			Name:        "Boaty McBoatface",
			Number:      "4200000000000042",
		},
		PaymentType: checkouts.ProcessCheckoutBodyPaymentTypeCard,
	})
	if err != nil {
		log.Printf("[ERROR] process checkout: %v", err)
		return
	}

	if accepted, ok := checkoutSuccess.AsCheckoutSuccess(); ok {
		log.Printf("[INFO] checkout success: id=%q, transaction_id=%q", *accepted.Id, *(*accepted.Transactions)[0].Id)
	}

	if accepted, ok := checkoutSuccess.AsCheckoutAccepted(); ok {
		log.Printf("[INFO] checkout accepted: redirect_to=%q", *accepted.NextStep.RedirectUrl)
	}
}
