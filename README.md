<div align="center">

# sumup-go

[![Stars](https://img.shields.io/github/stars/sumup/sumup-go?style=social)](https://github.com/sumup/sumup-go/)
[![Go Reference](https://pkg.go.dev/badge/github.com/sumup/sumup-go.svg)](https://pkg.go.dev/github.com/sumup/sumup-go)
[![CI Status](https://github.com/sumup/sumup-go/workflows/CI/badge.svg)](https://github.com/sumup/sumup-go/actions/workflows/ci.yml)
[![License](https://img.shields.io/github/license/sumup/sumup-go)](./LICENSE)
[![Contributor Covenant](https://img.shields.io/badge/Contributor%20Covenant-v2.1%20adopted-ff69b4.svg)](https://github.com/sumup/sumup-go/tree/main/CODE_OF_CONDUCT.md)

</div>

_**IMPORTANT:** This SDK is under heavy development and subject to breaking changes._

The Golang SDK for the SumUp [API](https://developer.sumup.com).

To learn more, check out our [API Reference](https://developer.sumup.com/api) and [Documentation](https://developer.sumup.com).

## Installation

`sumup-go` is compatible with projects using Go Modules.

Import the SDK using:

```go
import (
	"github.com/sumup/sumup-go"
)
```

And run any of `go build`/`go install`/`go test` which will resolve the package automatically.

Alternatively, you can install the SDK using:

```bash
go get github.com/sumup/sumup-go
```

## Documentation

For complete documentation of SumUp APIs visit [developer.sumup.com](https://developer.sumup.com).
Alternatively, refer to this simple example to get started:

```go
package main

import (
	"context"
	"os"

	"github.com/sumup/sumup-go"
	"github.com/sumup/sumup-go/merchant"
)

func main() {
	client := sumup.NewClient()

	account, err := client.Merchant.GetAccount(context.Background(), merchant.GetAccountParams{})
	if err != nil {
		fmt.Printf("[ERROR] get merchant account: %v\n", err)
		return
	}

	fmt.Printf("[INFO] merchant profile: %+v\n", account.MerchantProfile)
}
```

## Authentication

The easiest form of authenticating with SumUp APIs is using [API keys](https://developer.sumup.com/docs/online-payments/introduction/authorization/#api-keys). You can create API keys in the [API key section](https://developer.sumup.com/protected/api-keys/) of the developer portal. Store them securely. The SDK by default loads the API key from `SUMUP_API_KEY` environment variable. Alternatively, provide API key on your own:

```go
client := sumup.NewClient(client.WithAPIKey("sup_sk_LZFWoLyd..."))
```

## Support

Our APIs and their public offering is limited and under heavy development. If you have any questions or inquiries reach out to our support team via the [Contact Form](https://developer.sumup.com/contact/).

For question specifically related to this Golang SDK please [Open an Issue](https://github.com/sumup/sumup-go/issues/new).

`sumup-go` SDK will always support latest 3 version of golang following the Golang [Release Policy](https://go.dev/doc/devel/release).
