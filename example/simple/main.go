package main

import (
	"context"
	"log"
	"os"

	"github.com/sumup/sumup-go"
)

func main() {
	client := sumup.NewClient().WithAuth(os.Getenv("SUMUP_KEY"))

	account, err := client.Merchant.Get(context.Background(), sumup.GetAccountParams{})
	if err != nil {
		log.Printf("[ERROR] get merchant account: %v", err)
		return
	}

	log.Printf("[INFO] merchant code: %s", *account.MerchantProfile.MerchantCode)
}
