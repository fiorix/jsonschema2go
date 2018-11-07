TEST_SCHEMA=testdata/nvd/nvd_cve_feed_json_1.0.schema

all: \
	binary \
	test \
	test-gen-go \
	test-gen-thrift

binary:
	go build -v

test:
	go test -v -cover

gen-golden: binary
	./jsonschema2go -gen go $(TEST_SCHEMA) > testdata/go.golden
	./jsonschema2go -gen thrift $(TEST_SCHEMA) > testdata/thrift.golden

test-gen-go: binary
	mkdir -p .test/go
	./jsonschema2go -gen go -o .test/go/schema.go -gofmt $(TEST_SCHEMA)
	(cd .test/go && go build -v)

test-gen-thrift: binary
	mkdir -p .test/thrift
	./jsonschema2go -gen thrift -o .test/thrift/schema.thrift $(TEST_SCHEMA)
	(cd .test/thrift && thrift -gen go schema.thrift)

clean:
	rm -rf .test jsonschema2go
