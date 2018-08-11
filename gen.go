package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/fiorix/jsonschema2go/schema"
)

// Gen open/fetch and parses JSON schema, and writes Go code to w.
func Gen(w io.Writer, pkgName, jsonSchemaURL string) error {
	fmt.Fprintf(w, "// Package %s was auto-generated.\n// Command: %s\n", pkgName, strings.Join(os.Args, " "))
	fmt.Fprintf(w, "package %s\n\n", pkgName)
	return fetchAndGenerate(w, jsonSchemaURL)
}

type schema2go struct {
	root      *schema.Schema
	baseURL   string
	rootType  string
	typeCache map[string]struct{}
}

func fetchAndGenerate(w io.Writer, schemaURL string) error {
	root, err := openAndDecodeSchema(schemaURL)
	if err != nil {
		return err
	}

	baseURL, file := filepath.Split(schemaURL)
	name := goPublicType(file)

	g := &schema2go{
		root:      root,
		baseURL:   baseURL,
		rootType:  name,
		typeCache: make(map[string]struct{}),
	}

	return g.genStruct(w, name, schemaURL, root.Type)
}

func openAndDecodeSchema(schemaURL string) (*schema.Schema, error) {
	f, err := openSourceFile(schemaURL)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var root schema.Schema
	err = json.NewDecoder(f).Decode(&root)
	if err != nil {
		return nil, err
	}

	return &root, nil
}

var goPublicTypeRegexp = regexp.MustCompile("[0-9A-Za-z]+")

func goPublicType(name string) string {
	parts := goPublicTypeRegexp.FindAllString(name, -1)

	name = ""
	for _, part := range parts {
		name += strings.Title(part)
	}

	return name
}

func openSourceFile(schemaURL string) (io.ReadCloser, error) {
	u, err := url.Parse(schemaURL)
	if err != nil {
		return nil, err
	}

	switch {
	case u.Host == "" && u.Path != "":
		return os.Open(u.Path)
	case u.Host != "" && u.Path != "":
		resp, err := http.Get(schemaURL)
		if err != nil {
			return nil, err
		}
		return resp.Body, nil
	default:
		return nil, fmt.Errorf("invalid input schema: %q", schemaURL)
	}
}

func (g *schema2go) genStruct(w io.Writer, name, src string, t *schema.Type) error {
	if !strings.HasPrefix(name, g.rootType) {
		name = g.rootType + name
	}

	if _, exists := g.typeCache[name]; exists {
		return nil
	}

	g.typeCache[name] = struct{}{}

	switch {
	case t.Type != "object":
		return fmt.Errorf("schema type is not object: %q (root=%v)", name, src != "")
	case len(t.Properties) == 0:
		return fmt.Errorf("schema type has no properties: %q (root=%v)", name, src != "")
	}

	fmt.Fprintf(w, "// %s was auto-generated.\n", name)
	if t.Description != "" {
		fmt.Fprintf(w, "// %s\n", strings.Replace(t.Description, "\n", " ", -1))
	}
	if src != "" {
		fmt.Fprintf(w, "//\n// Source: %s\n", src)
	}
	fmt.Fprintf(w, "type %s struct {\n", name)

	var b bytes.Buffer

	for fieldName, fieldProp := range t.Properties {
		fieldTag := fieldName
		fieldName = goPublicType(fieldName)
		fieldType, err := g.genStructField(&b, name, fieldName, fieldProp)
		if err != nil {
			return err
		}
		fmt.Fprintf(w, "\t%s %s `json:\"%s\"`\n", fieldName, fieldType, fieldTag)
	}

	fmt.Fprintf(w, "}\n\n")
	_, err := io.Copy(w, &b)

	return err
}

func (g *schema2go) genStructField(w io.Writer, parent, name string, t *schema.Type) (string, error) {
	if t.Ref != "" {
		return g.genStructFieldFromRef(w, parent, name, t)
	}

	switch t.Type {
	case "string":
		return "string", nil
	case "number":
		return "float64", nil // TODO: handle min/max, write helper funcs to w
	case "boolean":
		return "bool", nil
	case "array":
		elemType, err := g.genStructField(w, parent, name, t.Items)
		if err != nil {
			return "", err
		}
		return "[]" + elemType, nil
	case "object":
		if strings.HasSuffix(parent, name) {
			return /* "*" + */ parent, nil
		}
		typeName := goPublicType(parent + "_" + name)
		err := g.genStruct(w, typeName, "", t)
		if err != nil {
			return "", err
		}
		return /* "*" + */ typeName, nil
	default:
		if t.Enum == nil {
			return "", fmt.Errorf("unknown field type for %q: %q", name, t.Type)
		}
		return "string", nil // TODO: handle enum properly
	}
}

func (g *schema2go) genStructFieldFromRef(w io.Writer, parent, name string, t *schema.Type) (string, error) {
	u, err := url.Parse(t.Ref)
	if err != nil {
		return "", err
	}

	switch {
	case strings.HasPrefix(u.Fragment, "/definitions/"):
		key := u.Fragment[len("/definitions/"):]
		def, exists := g.root.Definitions[key]
		if !exists {
			return "", fmt.Errorf("unknown reference for %q: %q", name, t.Ref)
		}

		if def.Type == "" && len(def.Properties) > 0 {
			def.Type = "object"
		}

		parent = g.rootType
		name = goPublicType(key)
		return g.genStructField(w, parent, name, def)

	case u.Path != "":
		err := fetchAndGenerate(w, g.genURL(u.Host, u.Path))
		if err != nil {
			return "", err
		}

		name = goPublicType(u.Path)
		return /* "*" + */ name, nil
	}

	return "", nil
}

func (g *schema2go) genURL(host, path string) string {
	if host == "" {
		return g.baseURL + path
	}
	return host + path
}
