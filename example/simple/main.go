package main

import (
	"context"
	"os"

	"golang.org/x/exp/slog"

	"github.com/sumup/sumup-go"
)

func main() {
	client := sumup.NewClient().WithAuth(os.Getenv("SUMUP_KEY"))

	account, err := client.Merchant.Get(context.Background(), sumup.GetAccountParams{})
	if err != nil {
		slog.Error("get merchant account", slog.String("error", err.Error()))
		return
	}

	slog.Info("merchant account", slog.String("merchant_code", *account.MerchantProfile.MerchantCode))
}
