package sumup_test

import (
	"context"
	"log"
	"os"

	"github.com/sumup/sumup-go"
)

func ExampleClient() {
	client := sumup.NewClient()

	merchant, err := client.Merchants.Get(context.Background(), os.Getenv("SUMUP_MERCHANT_CODE"), sumup.MerchantsGetParams{})
	if err != nil {
		log.Printf("[ERROR] get merchant account: %v", err)
		return
	}

	log.Printf("[INFO] business profile: %+v", merchant.BusinessProfile)
}
