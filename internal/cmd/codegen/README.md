<div align="center">

# codegen

</div>

Code generator for [sumup-go](https://github.com/sumup/sumup-go).

## Go SDK

The `generate` command reads `openapi.json` and generates the Go client, services, request types, and response types in the repository root. Generate the SDK and refresh its API documentation from the repository root with:

```shell
make generate
```

## Go Code Samples

The `samples` command generates a deterministic, versioned JSON catalog of Go examples from the same intermediate representation used to generate the SDK. Each catalog entry contains a complete, formatted Go program. Named OpenAPI request examples produce separate entries.

Generate a catalog from the repository root with:

```shell
make generate-codesamples
```

The target writes `code-samples.json` in the repository root by default. Set `CODESAMPLES_OUT` to use another path. Every generated program is compiled by the codegen test suite. When an SDK release is published, the release workflow regenerates the catalog from that tag and opens or updates a pull request in `sumup/sumup-developer`; the generated JSON is not committed to this repository.
