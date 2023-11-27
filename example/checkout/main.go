package main

import (
	"context"
	"log"
	"os"

	"github.com/sumup/sumup-go"
)

func main() {
	ctx := context.Background()
	client := sumup.NewClient().WithAuth(os.Getenv("SUMUP_KEY"))

	checkout, err := client.Checkouts.Create(ctx, sumup.CreateCheckoutBody{
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

	checkoutSuccess, err := client.Checkouts.Process(ctx, *checkout.Id, sumup.ProcessCheckoutBody{
		Card: &sumup.Card{
			Cvv:         "123",
			ExpiryMonth: "12",
			ExpiryYear:  "2023",
			Name:        "Boaty McBoatface",
			Number:      "4200000000000042",
		},
		PaymentType: sumup.ProcessCheckoutBodyPaymentTypeCard,
	})
	if err != nil {
		log.Printf("[ERROR] process checkout: %v", err)
		return
	}

	log.Printf("[INFO] checkout processed: id=%q, transaction_id=%q", *checkoutSuccess.Id, string(*(*checkoutSuccess.Transactions)[0].Id))
}
