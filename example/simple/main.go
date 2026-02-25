package main

import (
	"context"
	"log"
	"os"

	"github.com/sumup/sumup-go"
)

func main() {
	client := sumup.NewClient()

	account, err := client.Merchants.Get(context.Background(), os.Getenv("SUMUP_MERCHANT_CODE"), sumup.MerchantsGetParams{})
	if err != nil {
		log.Printf("[ERROR] get merchant account: %v", err)
		return
	}

	log.Printf("[INFO] merchant code: %s", account.MerchantCode)
}
