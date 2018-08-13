all: \
	binary \
	test \
	test-gen-go \
	test-gen-thrift

binary:
	go build -v

test:
	go test -v -cover

gen-golden:
	./jsonschema2go -gen go testdata/nvd/nvd_cve_feed_json_0.1_beta.schema > testdata/go.golden
	./jsonschema2go -gen thrift testdata/nvd/nvd_cve_feed_json_0.1_beta.schema > testdata/thrift.golden

test-gen-go:
	mkdir -p .test/go
	./jsonschema2go -gen go -o .test/go/schema.go -gofmt ./testdata/nvd/nvd_cve_feed_json_0.1_beta.schema
	(cd .test/go && go build -v)

test-gen-thrift:
	mkdir -p .test/thrift
	./jsonschema2go -gen thrift -o .test/thrift/schema.thrift ./testdata/nvd/nvd_cve_feed_json_0.1_beta.schema
	(cd .test/thrift && thrift -gen go schema.thrift)

clean:
	rm -rf .test jsonschema2go
