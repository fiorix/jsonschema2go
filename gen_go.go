package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strings"
)

var (
	flagGoPkg = flag.String("gopkg", "schema", "set name of generated go package")
	flagGoPtr = flag.Bool("goptr", false, "generate go struct fields as pointers")
	flagGoFmt = flag.Bool("gofmt", false, "run gofmt on generated go code")
)

var goPublicTypeRegexp = regexp.MustCompile("[0-9A-Za-z]+")

func goPublicType(name string) string {
	parts := goPublicTypeRegexp.FindAllString(name, -1)

	name = ""
	for _, part := range parts {
		name += strings.Title(part)
	}

	return name
}

var goTypeConv = typeConv{
	"boolean": "bool",
	"number":  "float64",
	"string":  "string",
}

// GenGo generates Go code to w.
func GenGo(w io.Writer, jsonSchemaURL string) error {
	list, err := ParseSchema(jsonSchemaURL)
	if err != nil {
		return err
	}

	pkg := *flagGoPkg
	var b bytes.Buffer
	fmt.Fprintf(&b, "// Package %s was auto-generated.\n", pkg)
	fmt.Fprintf(&b, "// Command: %s\n", commandLine())
	fmt.Fprintf(&b, "package %s\n\n", pkg)

	for _, structType := range list {
		structName := goPublicType(structType.Name)
		fmt.Fprintf(&b, "// %s was auto-generated.\n", structName)
		if structType.Desc != "" {
			if !strings.HasSuffix(structType.Desc, ".") {
				structType.Desc += "."
			}
			fmt.Fprintf(&b, "// %s\n", structType.Desc)
		}
		if structType.File != "" {
			fmt.Fprintf(&b, "// Source: %s\n", structType.File)
		}
		fmt.Fprintf(&b, "type %s struct {\n", structName)

		fields := make([]string, 0, len(structType.Fields))

		for _, structField := range structType.Fields {
			var line strings.Builder
			fmt.Fprintf(&line, "\t%s ", goPublicType(structField.Name))
			if structField.IsSlice {
				fmt.Fprintf(&line, "[]")
			}
			if structField.IsStruct {
				if *flagGoPtr {
					fmt.Fprintf(&line, "*")
				}
				io.WriteString(&line, goPublicType(structField.Type))
			} else {
				io.WriteString(&line, goTypeConv.Get(structField.Type))
			}
			fmt.Fprintf(&line, " `json:\"%s\"`", structField.Name)
			fields = append(fields, line.String())
		}

		sort.Strings(fields)
		io.WriteString(&b, strings.Join(fields, "\n"))
		io.WriteString(&b, "\n}\n\n")
	}

	if !*flagGoFmt {
		_, err := io.Copy(w, &b)
		return err
	}

	cmd := exec.Command("gofmt")
	cmd.Stdin = &b
	cmd.Stdout = w
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
