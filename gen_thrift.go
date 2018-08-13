package main

import (
	"flag"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"
)

var (
	flagThriftNS = flag.String("thriftns", "go schema,py schema", "set comma separated list of thrift namespaces")
)

var thriftPublicTypeRegexp = regexp.MustCompile("[0-9A-Za-z]+")

func thriftPublicType(name string) string {
	parts := thriftPublicTypeRegexp.FindAllString(name, -1)
	return strings.Join(parts, "_")
}

//var thriftPublicType = goPublicType

var thriftTypeConv = typeConv{
	"boolean": "bool",
	"number":  "double",
	"string":  "string",
}

// GenThrift generates thrift spec to w.
func GenThrift(w io.Writer, jsonSchemaURL string) error {
	list, err := ParseSchema(jsonSchemaURL)
	if err != nil {
		return err
	}

	fmt.Fprintf(w, "// This thrift spec was auto-generated.\n")
	fmt.Fprintf(w, "// Command: %s\n\n", commandLine())

	for _, ns := range strings.Split(*flagThriftNS, ",") {
		fmt.Fprintf(w, "namespace %s\n", strings.TrimSpace(ns))
	}
	io.WriteString(w, "\n")

	for _, structType := range list {
		//structName := thriftPublicType(structType.Name)
		structName := goPublicType(structType.Name)
		fmt.Fprintf(w, "// %s was auto-generated.\n", structName)
		if structType.Desc != "" {
			if !strings.HasSuffix(structType.Desc, ".") {
				structType.Desc += "."
			}
			fmt.Fprintf(w, "// %s\n", structType.Desc)
		}
		if structType.File != "" {
			fmt.Fprintf(w, "// Source: %s\n", structType.File)
		}
		fmt.Fprintf(w, "struct %s {\n", structName)

		fields := make(map[string]string)

		for _, structField := range structType.Fields {
			var t string
			if structField.IsStruct {
				//t = thriftPublicType(structField.Type)
				t = goPublicType(structField.Type)
			} else {
				t = thriftTypeConv.Get(structField.Type)
			}

			var line strings.Builder

			if structField.IsSlice {
				fmt.Fprintf(&line, "list<%s>", t)
			} else {
				io.WriteString(&line, t)
			}

			name := thriftPublicType(structField.Name)
			fmt.Fprintf(&line, " %s;", name)

			fields[name] = line.String()
		}

		keys := make([]string, 0, len(fields))
		for key := range fields {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		for i, key := range keys {
			fmt.Fprintf(w, " %d: %s\n", i+1, fields[key])
		}

		fmt.Fprintf(w, "}\n\n")
	}

	return nil
}
