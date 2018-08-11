package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/fiorix/jsonschema2go/testdata"
)

func TestGen(t *testing.T) {
	pkgName := "schema"
	srcFile := "./testdata/nvd/nvd_cve_feed_json_0.1_beta.schema"
	err := Gen(ioutil.Discard, pkgName, srcFile)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGenDecodeEncode(t *testing.T) {
	want := testdata.NvdCveFeedJson01BetaSchema{
		CVEDataType:         "CVE",
		CVEDataFormat:       "MITRE",
		CVEDataVersion:      "4.0",
		CVEDataNumberOfCVEs: "1",
		CVEDataTimestamp:    "2018-08-10T22:00Z",
		CVEItems: []testdata.NvdCveFeedJson01BetaSchemaDefCveItem{
			testdata.NvdCveFeedJson01BetaSchemaDefCveItem{
				Cve: testdata.CVEJSON40MinSchema{
					DataType:    "CVE",
					DataFormat:  "MITRE",
					DataVersion: "4.0",
					CVEDataMeta: testdata.CVEJSON40MinSchemaCVEDataMeta{
						ID:       "CVE-2011-4181",
						ASSIGNER: "cve@mitre.org",
					},
					Affects: testdata.CVEJSON40MinSchemaAffects{
						Vendor: testdata.CVEJSON40MinSchemaAffectsVendor{
							VendorData: []testdata.CVEJSON40MinSchemaAffectsVendorVendorData{
								{
									VendorName: "opensuse",
									Product: testdata.CVEJSON40MinSchemaAffectsVendorVendorDataProduct{
										ProductData: []testdata.CVEJSON40MinSchemaProduct{
											{
												ProductName: "open_build_service",
												Version: testdata.CVEJSON40MinSchemaProductVersion{
													VersionData: []testdata.CVEJSON40MinSchemaProductVersionVersionData{
														{
															VersionAffected: "1.7",
															VersionValue:    "1.7.0",
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
					Problemtype: testdata.CVEJSON40MinSchemaProblemtype{
						ProblemtypeData: []testdata.CVEJSON40MinSchemaProblemtypeProblemtypeData{
							{
								Description: []testdata.CVEJSON40MinSchemaLangString{
									{
										Lang:  "en",
										Value: "CWE-20",
									},
								},
							},
						},
					},
					References: testdata.CVEJSON40MinSchemaReferences{
						ReferenceData: []testdata.CVEJSON40MinSchemaReference{
							{
								Url:       "https://bugzilla.suse.com/show_bug.cgi?id=734003",
								Name:      "https://bugzilla.suse.com/show_bug.cgi?id=734003",
								Refsource: "CONFIRM",
								Tags:      []string{"Issue Tracking"},
							},
						},
					},
					Description: testdata.CVEJSON40MinSchemaDescription{
						DescriptionData: []testdata.CVEJSON40MinSchemaLangString{
							{
								Lang:  "en",
								Value: "A vulnerability in open build service allows remote attackers to gain access to source files even though source access is disabled. Affected releases are SUSE open build service up to and including version 2.1.15 (for 2.1) and before version 2.3.",
							},
						},
					},
				},
				Configurations: testdata.NvdCveFeedJson01BetaSchemaDefConfigurations{
					CVEDataVersion: "4.0",
					Nodes: []testdata.NvdCveFeedJson01BetaSchemaDefNode{
						{
							Operator: "OR",
							Cpe: []testdata.NvdCveFeedJson01BetaSchemaDefNodeCpe{
								{
									Vulnerable:          true,
									Cpe23Uri:            "cpe:2.3:a:opensuse:open_build_service:*:*:*:*:*:*:*:*",
									VersionEndExcluding: "2.3",
								},
							},
						},
					},
				},
				Impact: testdata.NvdCveFeedJson01BetaSchemaDefImpact{
					BaseMetricV3: testdata.NvdCveFeedJson01BetaSchemaDefImpactBaseMetricV3{
						ExploitabilityScore: 3.9,
						ImpactScore:         3.6,
						CvssV3: testdata.CvssV30Json{
							Version:            "3.0",
							VectorString:       "CVSS:3.0/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:N/A:N",
							AttackVector:       "NETWORK",
							AttackComplexity:   "LOW",
							PrivilegesRequired: "NONE",
							UserInteraction:    "NONE",
							Scope:              "UNCHANGED",
							ConfidentialityImpact: "HIGH",
							IntegrityImpact:       "NONE",
							AvailabilityImpact:    "NONE",
							BaseScore:             7.5,
							BaseSeverity:          "HIGH",
						},
					},
				},
				PublishedDate:    "2018-06-11T15:29Z",
				LastModifiedDate: "2018-08-02T12:52Z",
			},
		},
	}

	f, err := os.Open("testdata/cve-golden.json")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	var have testdata.NvdCveFeedJson01BetaSchema
	err = json.NewDecoder(f).Decode(&have)
	if err != nil {
		t.Fatal(err)
	}

	wantItems, haveItems := want.CVEItems, have.CVEItems
	want.CVEItems, have.CVEItems = nil, nil

	t.Run("root", func(t *testing.T) {
		if !reflect.DeepEqual(want, have) {
			t.Fatalf("invalid json:\nwant:\n%#v\nhave:\n%#v\n", want, have)
		}
	})

	t.Run("cve_items", func(t *testing.T) {
		if !reflect.DeepEqual(wantItems, haveItems) {
			t.Fatalf("invalid json:\nwant:\n%#v\nhave:\n%#v\n", wantItems, haveItems)
		}
	})
}
