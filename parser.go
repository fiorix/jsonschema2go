package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/fiorix/jsonschema2go/schema"
)

// ParseSchema open/fetch and parses JSON schema.
func ParseSchema(jsonSchemaURL string) (StructList, error) {
	var list StructList
	err := fetchAndGenerate(&list, jsonSchemaURL)
	if err != nil {
		return nil, err
	}
	return list, nil
}

// StructList is a list of struct types (JSON objects) from the schema.
type StructList []structType

type structType struct {
	File     string
	Name     string
	Desc     string
	Fields   []structField
	Required map[string]struct{}
}

type structField struct {
	Name     string
	Type     string
	IsStruct bool
	IsSlice  bool
	// TODO: min/max number and array items, enum options
}

type schema2go struct {
	root      *schema.Schema
	baseURL   string
	rootType  string
	typeCache map[string]struct{}
}

var nameExtRegexp = regexp.MustCompile(`[._]min[.](json|schema)([.]gz)?$`)

func nameFromFile(file string) string {
	idx := nameExtRegexp.FindStringIndex(file)
	if len(idx) > 0 {
		return strings.Replace(file[0:idx[0]], ".", "", -1)
	}
	ext := path.Ext(file)
	file = strings.Replace(file, ".", "", -1)
	if ext == "" {
		return file
	}
	return file[0 : len(file)-len(ext)+1]
}

func fetchAndGenerate(list *StructList, schemaURL string) error {
	root, err := openAndDecodeSchema(schemaURL)
	if err != nil {
		return err
	}

	baseURL, file := filepath.Split(schemaURL)

	g := &schema2go{
		root:      root,
		baseURL:   baseURL,
		rootType:  nameFromFile(file),
		typeCache: make(map[string]struct{}),
	}

	return g.genStruct(list, g.rootType, schemaURL, root.Type)
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

func (g *schema2go) genStruct(list *StructList, name, file string, t *schema.Type) error {
	if _, exists := g.typeCache[name]; exists {
		return nil
	}

	g.typeCache[name] = struct{}{}

	switch {
	case t.Type != "object":
		return fmt.Errorf("schema type is not object: %q (root=%v)", name, file != "")
	case len(t.Properties) == 0:
		return fmt.Errorf("schema type has no properties: %q (root=%v)", name, file != "")
	}

	st := structType{
		Name:     name,
		File:     file,
		Desc:     t.Description,
		Required: make(map[string]struct{}),
	}

	keys := make([]string, 0, len(t.Properties))
	for prop := range t.Properties {
		keys = append(keys, prop)
	}

	sort.Strings(keys)

	for _, fieldName := range keys {
		fieldProp := t.Properties[fieldName]
		field, err := g.genStructField(list, name, fieldName, fieldProp)
		if err != nil {
			return err
		}
		field.Name = fieldName
		st.Fields = append(st.Fields, *field)
	}

	for _, fieldName := range t.Required {
		st.Required[fieldName] = struct{}{}
	}

	*list = append(*list, st)

	return nil
}

type typeConv map[string]string

func (t typeConv) Get(s string) string {
	n, ok := t[s]
	if !ok {
		panic("unsupported type: " + s)
	}
	return n
}

func (g *schema2go) genStructField(list *StructList, parent, name string, t *schema.Type) (*structField, error) {
	if t.Ref != "" {
		return g.genStructFieldFromRef(list, parent, name, t)
	}

	switch t.Type {
	case "string", "number", "boolean":
		// TODO: handle number's min/max, etc
		f := &structField{
			Name: name,
			Type: t.Type,
		}
		return f, nil
	case "array":
		f, err := g.genStructField(list, parent, name, t.Items)
		if err != nil {
			return nil, err
		}
		f.IsSlice = true
		return f, nil
	case "object":
		f := &structField{
			Name:     name,
			Type:     path.Join(parent, name),
			IsStruct: true,
		}
		err := g.genStruct(list, f.Type, "", t)
		if err != nil {
			return nil, err
		}
		return f, nil
	default:
		if t.Enum == nil {
			return nil, fmt.Errorf("unknown field type for %q: %q", name, t.Type)
		}
		f := &structField{
			Name: name,
			Type: "string",
		}
		return f, nil // TODO: handle enum properly
	}
}

func (g *schema2go) genStructFieldFromRef(list *StructList, parent, name string, t *schema.Type) (*structField, error) {
	u, err := url.Parse(t.Ref)
	if err != nil {
		return nil, err
	}

	switch {
	case strings.HasPrefix(u.Fragment, "/definitions/"):
		key := u.Fragment[len("/definitions/"):]
		def, exists := g.root.Definitions[key]
		if !exists {
			return nil, fmt.Errorf("unknown reference for %q: %q", name, t.Ref)
		}

		if def.Type == "" && len(def.Properties) > 0 {
			def.Type = "object"
		}

		parent, name = g.rootType, key
		f, err := g.genStructField(list, parent, name, def)
		if err != nil {
			return nil, err
		}
		f.Name = key
		return f, nil

	case u.Path != "":
		err := fetchAndGenerate(list, g.genURL(u.Host, u.Path))
		if err != nil {
			return nil, err
		}

		f := &structField{
			Name:     name,
			Type:     nameFromFile(u.Path),
			IsStruct: true,
		}

		return f, nil
	}

	panic("unsupported $ref: " + t.Ref)
}

func (g *schema2go) genURL(host, path string) string {
	if host == "" {
		return g.baseURL + path
	}
	return host + path
}

func commandLine() string {
	args := make([]string, 0, len(os.Args))
	for _, arg := range os.Args {
		if strings.Index(arg, " ") == -1 || arg[0] != '-' {
			args = append(args, arg)
			continue
		}

		pos := strings.Index(arg, "=")
		pos++
		arg = arg[0:pos] + fmt.Sprintf("%q", arg[pos:])
		args = append(args, arg)
	}

	return strings.Join(args, " ")
}
