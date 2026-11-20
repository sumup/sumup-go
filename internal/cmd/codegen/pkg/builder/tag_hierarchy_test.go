package builder

import (
	"strings"
	"testing"

	"github.com/pb33f/libopenapi"
)

func TestNestedReaderResources(t *testing.T) {
	t.Parallel()

	doc, err := libopenapi.NewDocument([]byte(`
openapi: 3.1.0
info: {title: Readers, version: 1.0.0}
tags:
  - name: Readers
  - name: ReaderCheckouts
    x-displayName: Checkouts
x-tagGroups:
  - name: Readers
    tags: [ReaderCheckouts]
paths:
  /readers/checkouts:
    get:
      operationId: ListReaders
      tags: [Readers]
      responses:
        '204': {description: OK}
    post:
      operationId: CreateReaderCheckout
      tags: [ReaderCheckouts]
      x-codegen: {method_name: create_checkout}
      parameters:
        - in: header
          name: X-SumUp-Merchant-Code
          required: false
          schema: {type: string}
      requestBody:
        content:
          application/json:
            schema: {type: object, properties: {reference: {type: string}}}
      responses:
        '204': {description: OK}
components: {schemas: {}}
`))
	if err != nil {
		t.Fatal(err)
	}
	model, err := doc.BuildV3Model()
	if err != nil {
		t.Fatal(err)
	}

	builder := New(Config{Out: t.TempDir()})
	if err := builder.Load(&model.Model); err != nil {
		t.Fatal(err)
	}
	child := builder.pathsByTag["readercheckouts"].PathItems.GetOrZero("/readers/checkouts")
	if child.Get != nil || child.Post == nil {
		t.Fatal("the child must only contain its own operation")
	}
	parent := builder.pathsByTag["readers"].PathItems.GetOrZero("/readers/checkouts")
	if parent.Get == nil || parent.Post == nil {
		t.Fatal("the parent must still contain both operations")
	}
	if path := builder.resourcePath("readercheckouts"); path != "Readers.Checkouts" {
		t.Fatalf("wrong resource path %q", path)
	}
	for _, tag := range []string{"readers", "readercheckouts"} {
		method, err := builder.operationToMethod(tag, "POST", "/readers/checkouts", child.Post)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(method.ParamsString(), "body ReadersCreateCheckoutParams, params ...ReadersCreateCheckoutOptions") {
			t.Errorf("unexpected checkout method signature for %s: %s", tag, method.ParamsString())
		}
	}
	fields := builder.pathsToParamTypes("readers", builder.pathsByTag["readers"])
	if len(fields) != 3 || !strings.Contains(fields[2].String(), `headers.Set("X-SumUp-Merchant-Code", *p.XSumUpMerchantCode)`) {
		t.Fatalf("expected checkout options to set the merchant header: %v", fields)
	}
}
