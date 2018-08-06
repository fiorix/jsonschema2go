// Package gen is part of jsonschema2go and provides a Go code generator from JSON schema.
package gen

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

// Go open/fetch and parses JSON schema, and writes Go code to w.
func Go(w io.Writer, pkgName, jsonSchemaURL string) error {
	fmt.Fprintf(w, "// Package %s was auto-generated.\n", pkgName)
	fmt.Fprintf(w, "package %s\n\n", pkgName)
	return fetchAndGenerate(w, jsonSchemaURL)
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

type schema2go struct {
	root      *schema.Schema
	baseURL   string
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
		typeCache: make(map[string]struct{}),
	}

	return g.genObjectType(w, name, schemaURL, root.Type)
}

func openAndDecodeSchema(schemaURL string) (*schema.Schema, error) {
	f, err := getSourceFile(schemaURL)
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

func getSourceFile(schemaURL string) (io.ReadCloser, error) {
	u, err := url.Parse(schemaURL)
	if err != nil {
		return nil, err
	}

	switch {
	case u.Host == "" && u.Path != "":
		return os.Open(u.Path)
	case u.Host != "" && u.Path != "":
		resp, err := http.Get(u.String())
		if err != nil {
			return nil, err
		}
		return resp.Body, nil
	default:
		return nil, fmt.Errorf("invalid input schema: %q", schemaURL)
	}
}

func (g *schema2go) genObjectType(w io.Writer, name, src string, t *schema.Type) error {
	switch {
	case t.Type != "object":
		return fmt.Errorf("root schema type is not object: %q", name)
	case len(t.Properties) == 0:
		return fmt.Errorf("root schema type has no properties: %q", name)

	}

	if _, exists := g.typeCache[name]; exists {
		return nil
	}

	g.typeCache[name] = struct{}{}

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
		fieldType, err := g.genFieldType(&b, fieldName, fieldProp)
		if err != nil {
			return err
		}
		fmt.Fprintf(w, "\t%s %s `json:\"%s\"`\n", fieldName, fieldType, fieldTag)
	}

	fmt.Fprintf(w, "}\n\n")
	_, err := io.Copy(w, &b)

	return err
}

func (g *schema2go) genFieldType(w io.Writer, name string, t *schema.Type) (string, error) {
	if t.Ref != "" {
		return g.genFieldTypeFromRef(w, name, t)
	}

	switch t.Type {
	case "string":
		return "string", nil
	case "number":
		return "float64", nil // TODO: handle min/max, write helper funcs to w
	case "boolean":
		return "bool", nil
	case "array":
		elemType, err := g.genFieldType(w, name, t.Items)
		if err != nil {
			return "", err
		}
		return "[]" + elemType, nil
	case "object":
		err := g.genObjectType(w, name, "", t)
		if err != nil {
			return "", err
		}
		return "*" + name, nil
	default:
		if t.Enum == nil {
			return "", fmt.Errorf("unknown field type for %q: %q", name, t.Type)
		}
		return "string", nil // TODO: handle enum properly
	}
}

func (g *schema2go) genFieldTypeFromRef(w io.Writer, name string, t *schema.Type) (string, error) {
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

		name = goPublicType(key)
		return g.genFieldType(w, name, def)

	case u.Path != "":
		err := fetchAndGenerate(w, g.genURL(u.Host, u.Path))
		if err != nil {
			return "", err
		}

		name = goPublicType(u.Path)
		return "*" + name, nil
	}

	return "", nil
}

func (g *schema2go) genURL(host, path string) string {
	if host == "" {
		return g.baseURL + path
	}
	return host + path
}
