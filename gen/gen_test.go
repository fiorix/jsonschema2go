package gen

import (
	"io/ioutil"
	"testing"
)

func TestGo(t *testing.T) {
	pkgName := "schema"
	srcFile := "./testdata/nvd_cve_feed_json_0.1_beta.schema"
	err := Go(ioutil.Discard, pkgName, srcFile)
	if err != nil {
		t.Fatal(err)
	}
}
