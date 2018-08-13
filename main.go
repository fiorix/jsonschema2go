package main

import (
	"flag"
	"fmt"
	"io"
	"os"
)

func main() {
	flag.Usage = func() {
		fmt.Printf("use: %s [flags] schema.json\n", os.Args[0])
		fmt.Printf("The input schema may be a local file or URL.\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	outputFile := flag.String("o", "-", "set name of output file")
	outputType := flag.String("gen", "go", "set output format: go, thrift")
	flag.Parse()

	if len(flag.Args()) == 0 {
		flag.Usage()
	}

	f, err := createFile(*outputFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer f.Close()

	jsonSchema := flag.Arg(0)

	switch *outputType {
	case "go":
		err = GenGo(f, jsonSchema)
	case "thrift":
		err = GenThrift(f, jsonSchema)
	default:
		err = fmt.Errorf("unsupported output type: %q", *outputType)
	}

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func createFile(name string) (io.WriteCloser, error) {
	switch name {
	case "", "-":
		return os.Stdout, nil
	default:
		return os.Create(name)
	}
}
