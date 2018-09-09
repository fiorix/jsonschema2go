# jsonschema2go

[![Build Status](https://secure.travis-ci.org/fiorix/jsonschema2go.png)](http://travis-ci.org/fiorix/jsonschema2go)

jsonschema2go is a code generator for JSON schemas. Supports schemas from local files or URL, and generates Go code, or thrift spec.

This is a very naive and incomplete implementation. I wrote this code specifically to codegen the [NVD JSON schema](https://nvd.nist.gov/vuln/data-feeds#JSON_FEED), based on a few requirements:

* The output is a single file
* Consistent output given same input
* Capable of generating at least Go and Thrift

### Download, install

Requires Go 1.10 or newer. The generated thrift spec requires thrift compiler 0.11 or newer.

Assuming you have a working Go environment:

```
go get github.com/fiorix/jsonschema2go
go install github.com/fiorix/jsonschema2go
```

Output binary is $GOPATH/bin/jsonschema2go.

### Usage

Generate Go code:

```
jsonschema2go -gen go https://csrc.nist.gov/schema/nvd/feed/0.1/nvd_cve_feed_json_0.1_beta.schema
```

Generate Thrift spec:

```
jsonschema2go -gen thrift https://csrc.nist.gov/schema/nvd/feed/0.1/nvd_cve_feed_json_0.1_beta.schema
```