<div align="center">

# sumup-go

[![Go Reference](https://pkg.go.dev/badge/github.com/sumup/sumup-go.svg)](https://pkg.go.dev/github.com/sumup/sumup-go)
[![CI Status](https://github.com/sumup/sumup-go/workflows/CI/badge.svg)](https://github.com/sumup/sumup-go/actions/workflows/ci.yml)

</div>

_**IMPORTANT:** This SDK is under heavy development and subject to breaking changes._

The Golang SDK for the SumUp [API](https://developer.sumup.com).

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
)

func main() {
	client := sumup.NewClient().WithAuth(os.Getenv("SUMUP_KEY"))

	account, err := client.Merchant.GetAccount(context.Background(), sumup.GetAccountParams{})
	if err != nil {
		fmt.Printf("get merchant account: %s", err.Error())
		return
	}

	fmt.Printf("authorized for merchant %q", *account.MerchantProfile.MerchantCode)
}
```

## Authentication

The easiest form of authenticating with SumUp APIs is using [API keys](https://developer.sumup.com/docs/online-payments/introduction/authorization/#api-keys). You can create API keys in the [API key section](https://developer.sumup.com/protected/api-keys/) of the developer portal. Store them securely and use them with:

```go
client := sumup.NewClient().WithAuth(os.Getenv("SUMUP_KEY"))
```

## Support

Our APIs and their public offering is limited and under heavy development. If you have any questions or inquiries reach out to our support team via the [Contact Form](https://developer.sumup.com/contact/).

For question specifically related to this Golang SDK please use the [Discussions](https://github.com/sumup/sumup-go/discussions) or [Open an Issue](https://github.com/sumup/sumup-go/issues/new).

`sumup-go` SDK will always support latest 3 version of golang following the Golang [Release Policy](https://go.dev/doc/devel/release).
