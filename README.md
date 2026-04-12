<div align="center">

# SumUp Go SDK

[![Stars](https://img.shields.io/github/stars/sumup/sumup-go)](https://github.com/sumup/sumup-go/)
[![Go Reference](https://pkg.go.dev/badge/github.com/sumup/sumup-go.svg)](https://pkg.go.dev/github.com/sumup/sumup-go)
[![Documentation][docs-badge]](https://developer.sumup.com)
[![CI Status](https://github.com/sumup/sumup-go/workflows/CI/badge.svg)](https://github.com/sumup/sumup-go/actions/workflows/ci.yml)
[![License](https://img.shields.io/github/license/sumup/sumup-go)](./LICENSE)

</div>

_**IMPORTANT:** This SDK is under development. We might still introduce minor breaking changes before reaching v1._

The [Go](https://go.dev/) SDK for the SumUp [API](https://developer.sumup.com).

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
	"log"

	"github.com/sumup/sumup-go"
)

func main() {
	client := sumup.NewClient()

	merchant, err := client.Merchants.Get(context.Background(), "MCNPLE22", sumup.MerchantsGetParams{})
	if err != nil {
		log.Printf("[ERROR] get merchant account: %v", err)
		return
	}

	log.Printf("[INFO] merchant code: %s", merchant.MerchantCode)
}
```

## Authentication

The easiest form of authenticating with SumUp APIs is using [API keys](https://developer.sumup.com/docs/online-payments/introduction/authorization/#api-keys). You can create API keys in the [API key section](https://developer.sumup.com/protected/api-keys/) of the developer portal. Store them securely. The SDK by default loads the API key from `SUMUP_API_KEY` environment variable. Alternatively, provide API key on your own:

```go
client := sumup.NewClient(client.WithAPIKey("sup_sk_LZFWoLyd..."))
```

## Events

Receive signed event notifications with callbacks generated from the OpenAPI spec:

```go
handler, err := client.EventsHandler(os.Getenv("SUMUP_EVENT_SECRET"), handleUnhandled)
if err != nil {
	return err
}
if err := handler.OnMemberUpdated(handleMemberUpdated); err != nil {
	return err
}
if err := handler.OnReaderCreated(handleReaderCreated); err != nil {
	return err
}

// Inside your HTTP handler, limit and read the raw body before dispatch.
r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
body, err := io.ReadAll(r.Body)
if err != nil {
	http.Error(w, "invalid event body", http.StatusBadRequest)
	return
}
err = handler.Handle(r.Context(), body, r.Header.Get(sumup.EventSignatureHeader))
// Return a 2xx response only after success.
```

Callbacks accept a context and their specific event type:

```go
func handleMemberUpdated(ctx context.Context, event *sumup.MemberUpdatedEvent) error {
	member, err := event.FetchObject(ctx)
	if err != nil {
		return err
	}
	log.Printf("member updated: %s", member.ID)
	return nil
}
```

The required fallback has signature `func(context.Context, sumup.EventNotification) error`.
It receives known events without a registered callback and unknown event types.
Registering a nil or duplicate callback returns an error; registration and handling are concurrency-safe.
Callbacks execute synchronously and must synchronize any shared state themselves.

Signatures use `X-SumUp-Webhook-Signature: t=<unix timestamp>,v1=<hex HMAC>`.
Verification happens before JSON parsing, uses the exact raw body, and accepts a fixed five minutes of clock skew. An empty signing secret is rejected.
The SDK accepts already-read bytes and leaves body-size limits to your HTTP receiver. Use `http.MaxBytesReader` before reading the body, as in the example above.

For manual dispatch, use `client.ParseEventNotification(secret, body, signature)` and a type switch.
Only use `ParseEventNotificationWithoutVerification` for fixtures or notifications already verified by trusted infrastructure.

`FetchObject(ctx)` returns the latest typed resource using the SDK's authentication and transport.
Unknown events fetch raw JSON. Resource URLs must use the default API origin (`https://api.sumup.com`).
URL credentials and fragments are ignored; the path and query use the client's configured base URL, authentication, and redirect policy.
Structured API errors are returned as `*sumup.Problem`, accessible with `errors.As`.
Deleted resources may no longer be fetchable. Treat events as notifications, make processing idempotent using the event ID,
and allow for retries and out-of-order delivery. Return a 5xx response for `*sumup.EventCallbackError` so failed processing can be retried.

See [the events example](./example/events) for a complete HTTP receiver and error handling.

## Examples

The repository includes several examples demonstrating different use cases:

**[simple](./example/simple)** - Basic merchant account information retrieval showing how to initialize the SDK and make a simple API call.
```sh
go run example/simple/main.go
```

**[checkout](./example/checkout)** - Creating and processing a checkout programmatically using test card details.
```sh
go run example/checkout/main.go
```

**[full](./example/full)** - Complete web application demonstrating the full checkout flow with the SumUp payment widget. Shows how to create checkouts and integrate the widget in a real application.
```sh
go run example/full/main.go
```
and visit http://localhost:8080

**[events](./example/events)** - HTTP server showing signature verification, typed callbacks, and resource fetching.
```sh
SUMUP_EVENT_SECRET=whsec_test go run example/events/main.go
```
Then send `POST` requests to http://localhost:8080/events

## Support

Our APIs and their public offering is limited and under heavy development. If you have any questions or inquiries reach out to our support team via the [Contact Form](https://developer.sumup.com/contact/).

For question specifically related to this Golang SDK please [Open an Issue](https://github.com/sumup/sumup-go/issues/new).

`sumup-go` SDK will always support latest 3 version of golang following the Golang [Release Policy](https://go.dev/doc/devel/release).

[docs-badge]: https://img.shields.io/badge/SumUp-documentation-white.svg?logo=data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMjQiIGhlaWdodD0iMjQiIHZpZXdCb3g9IjAgMCAyNCAyNCIgZmlsbD0ibm9uZSIgY29sb3I9IndoaXRlIiB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciPgogICAgPHBhdGggZD0iTTIyLjI5IDBIMS43Qy43NyAwIDAgLjc3IDAgMS43MVYyMi4zYzAgLjkzLjc3IDEuNyAxLjcxIDEuN0gyMi4zYy45NCAwIDEuNzEtLjc3IDEuNzEtMS43MVYxLjdDMjQgLjc3IDIzLjIzIDAgMjIuMjkgMFptLTcuMjIgMTguMDdhNS42MiA1LjYyIDAgMCAxLTcuNjguMjQuMzYuMzYgMCAwIDEtLjAxLS40OWw3LjQ0LTcuNDRhLjM1LjM1IDAgMCAxIC40OSAwIDUuNiA1LjYgMCAwIDEtLjI0IDcuNjlabTEuNTUtMTEuOS03LjQ0IDcuNDVhLjM1LjM1IDAgMCAxLS41IDAgNS42MSA1LjYxIDAgMCAxIDcuOS03Ljk2bC4wMy4wM2MuMTMuMTMuMTQuMzUuMDEuNDlaIiBmaWxsPSJjdXJyZW50Q29sb3IiLz4KPC9zdmc+
