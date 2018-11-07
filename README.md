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

```
use: ./jsonschema2go [flags] schema.json
The input schema may be a local file or URL.
  -gen string
    	set output format: go, thrift (default "go")
  -gofmt
    	run gofmt on generated go code
  -gopkg string
    	set name of generated go package (default "schema")
  -goptr
    	generate go struct fields as pointers
  -o string
    	set name of output file (default "-")
  -thriftns string
    	set comma separated list of thrift namespaces (default "go schema,py schema")
```

---

Generate Go code:

```
jsonschema2go -gen go https://csrc.nist.gov/schema/nvd/feed/1.0/nvd_cve_feed_json_1.0.schema
```

Generate Thrift spec:

```
jsonschema2go -gen thrift https://csrc.nist.gov/schema/nvd/feed/1.0/nvd_cve_feed_json_1.0.schema
```

### Generated type names

Naming is hard. In jsonschema2go I made the choice to name generated Go and Thrift types after their respective filenames. In JSON schema, the body of the document is a type, like a struct, and it has no name. Hence using the filename with some normalization to idiomatic Go, with an adjusted list of keywords (e.g. CVE) for initialisms.

You can see what it looks like in the golden files in the testdata directory.