package sumup_test

import (
	"context"
	"log"

	"github.com/sumup/sumup-go"
)

func ExampleClient() {
	client := sumup.NewClient()

	account, err := client.Merchant.Get(context.Background(), sumup.MerchantGetParams{})
	if err != nil {
		log.Printf("[ERROR] get merchant account: %v", err)
		return
	}

	log.Printf("[INFO] merchant profile: %+v", *account.MerchantProfile)
}
