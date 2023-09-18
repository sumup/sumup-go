package main

import (
	"context"
	"os"

	"golang.org/x/exp/slog"

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
		slog.Error("create checkout", slog.String("error", err.Error()))
		return
	}

	slog.Info("checkout created",
		slog.String("id", *checkout.Id),
		slog.Float64("amount", *checkout.Amount),
		slog.String("currency", string(*checkout.Currency)),
	)

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
		slog.Error("process checkout", slog.String("error", err.Error()))
		return
	}

	slog.Info("checkout processed",
		slog.String("id", *checkout.Id),
		slog.Float64("amount", *checkout.Amount),
		slog.String("currency", string(*checkout.Currency)),
		slog.String("transaction_id", string(*(*checkout.Transactions)[0].Id)),
	)
}
