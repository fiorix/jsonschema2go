package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/fiorix/jsonschema2go/gen"
)

func main() {
	flag.Usage = func() {
		fmt.Printf("use: %s [flags] schema.json\n", os.Args[0])
		fmt.Printf("The input schema may be a local file or URL.\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	pkgName := flag.String("p", "schema", "name of generated package")
	outputFile := flag.String("o", "-", "output file")
	flag.Parse()

	if len(flag.Args()) == 0 {
		flag.Usage()
	}

	out, err := createFile(*outputFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer out.Close()

	pr, pw := io.Pipe()

	gofmt := exec.Command("gofmt")
	gofmt.Stdin = pr
	gofmt.Stdout = out

	err = gofmt.Start()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = gen.Go(pw, *pkgName, flag.Arg(0))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	pw.Close()

	err = gofmt.Wait()
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
