# jsonschema2go

jsonschema2go is a Go code generator. It reads JSON schema and produces a single Go package with all types from the schema.
Supports local files and URLs.

This is a very naive and incomplete implementation. I wrote this code specifically to codegen the [NVD JSON schema](https://nvd.nist.gov/vuln/data-feeds#JSON_FEED).

### Usage

Example:

```
go run github.com/fiorix/jsonschema2go -o schema.go https://csrc.nist.gov/schema/nvd/feed/0.1/nvd_cve_feed_json_0.1_beta.schema
```
