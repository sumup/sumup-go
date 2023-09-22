// Package sumup provides a client for using the SumUp API.
/*
Usage:

	import "github.com/sumup/sumup-go"

Construct a new SumUp client, then use the various services on the client to
access different parts of the SumUp API. For example:

	client := sumup.NewClient().WithAuth(os.Getenv("SUMUP_KEY"))

	// get the account the client is currently authorized for
	account, err := client.Merchant.Get(context.Background(), sumup.GetAccountParams{})

The client is structured around individual services that correspond to the tags
in SumUp documentation https://developer.sumup.com/docs/api/sum-up-rest-api/.
*/
package sumup
