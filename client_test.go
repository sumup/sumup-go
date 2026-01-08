package sumup_test

import (
	"context"
	"log"

	"github.com/sumup/sumup-go"
	"github.com/sumup/sumup-go/merchant"
)

func ExampleClient() {
	client := sumup.NewClient()

	account, err := client.Merchant.Get(context.Background(), merchant.GetParams{})
	if err != nil {
		log.Printf("[ERROR] get merchant account: %v", err)
		return
	}

	log.Printf("[INFO] merchant profile: %+v", *account.MerchantProfile)
}
