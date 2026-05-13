package builder

import (
	"strings"
	"testing"

	"github.com/pb33f/libopenapi"
)

func TestBuildSamples(t *testing.T) {
	t.Parallel()

	spec := []byte(`
openapi: 3.0.3
info:
  title: test
  version: "1.0"
paths:
  /checkouts:
    post:
      operationId: create
      tags: [checkouts]
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CheckoutCreateRequest'
      responses:
        "200":
          description: ok
          content:
            application/json:
              schema:
                type: object
  /merchants/{merchant_code}/payment-methods:
    get:
      operationId: listAvailablePaymentMethods
      tags: [checkouts]
      parameters:
        - name: merchant_code
          in: path
          required: true
          schema:
            type: string
            example: MH4H92C7
        - name: amount
          in: query
          required: true
          schema:
            type: number
            format: float
            example: 10.1
        - name: currency
          in: query
          required: true
          schema:
            type: string
            example: EUR
      responses:
        "200":
          description: ok
          content:
            application/json:
              schema:
                type: object
components:
  schemas:
    CheckoutCreateRequest:
      type: object
      required:
        - checkout_reference
        - amount
        - currency
        - merchant_code
      properties:
        checkout_reference:
          type: string
          example: f00a8f74-b05d-4605-bd73-2a901bae5802
        amount:
          type: number
          format: float
          example: 10.1
        currency:
          type: string
          example: EUR
        merchant_code:
          type: string
          example: MH4H92C7
`)

	doc, err := libopenapi.NewDocument(spec)
	if err != nil {
		t.Fatalf("NewDocument() error = %v", err)
	}

	model, err := doc.BuildV3Model()
	if err != nil {
		t.Fatalf("BuildV3Model() error = %v", err)
	}

	b := New(Config{})
	if err := b.Load(&model.Model); err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	samples, err := b.BuildSamples()
	if err != nil {
		t.Fatalf("BuildSamples() error = %v", err)
	}

	create := samples["/checkouts"]["POST"].Sample
	if !strings.Contains(create, "client.Checkouts.Create(") {
		t.Fatalf("create sample missing method call: %s", create)
	}
	if !strings.Contains(create, `CheckoutReference: "f00a8f74-b05d-4605-bd73-2a901bae5802"`) {
		t.Fatalf("create sample missing checkout reference example: %s", create)
	}
	if !strings.Contains(create, `MerchantCode: "MH4H92C7"`) {
		t.Fatalf("create sample missing merchant code example: %s", create)
	}

	list := samples["/merchants/{merchant_code}/payment-methods"]["GET"].Sample
	if !strings.Contains(list, "client.Checkouts.ListAvailablePaymentMethods(") {
		t.Fatalf("list sample missing method call: %s", list)
	}
	if !strings.Contains(list, `"MH4H92C7"`) {
		t.Fatalf("list sample missing path parameter example: %s", list)
	}
	if !strings.Contains(list, `Amount: 10.1`) || !strings.Contains(list, `Currency: "EUR"`) {
		t.Fatalf("list sample missing query examples: %s", list)
	}
}
