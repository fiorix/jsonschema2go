package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"testing"
)

const testSchemaFile = "testdata/nvd/nvd_cve_feed_json_1.0.schema"

func TestParseSchema(t *testing.T) {
	_, err := ParseSchema(testSchemaFile)
	if err != nil {
		t.Fatal(err)
	}
}

type genFunc func(w io.Writer, jsonSchema string) error

var goldenCmdRegexp = regexp.MustCompile("// Command: .*\n")

func testGenEqualGolden(t *testing.T, f genFunc, file string) {
	var b bytes.Buffer
	err := f(&b, testSchemaFile)
	if err != nil {
		t.Fatal(err)
	}

	want, err := ioutil.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}

	// remove the command from generated code; won't ever match
	want = goldenCmdRegexp.ReplaceAll(want, nil)
	have := goldenCmdRegexp.ReplaceAll(b.Bytes(), nil)

	if !bytes.Equal(want, have) {
		t.Error("golden file (a) != generated file (b)")
		err = printDiff("gen", "tmp", want, have)
		if err != nil {
			t.Fatal(err)
		}
	}
}

// printDiff is a helper function used by tests.
func printDiff(prefix, ext string, a, b []byte) error {
	files := []struct {
		Name string
		Data []byte
	}{
		{Name: prefix + "-a." + ext, Data: a},
		{Name: prefix + "-b." + ext, Data: b},
	}
	for _, f := range files {
		defer os.Remove(f.Name)
		if err := ioutil.WriteFile(f.Name, f.Data, 0600); err != nil {
			return err
		}
	}
	var stdout, stderr bytes.Buffer
	cmd := exec.Command("diff", "-u", files[0].Name, files[1].Name)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("%v: %s", err, stdout.String())
	}
	return nil
}
