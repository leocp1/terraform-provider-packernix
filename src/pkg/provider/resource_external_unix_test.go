// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// Requires POSIX shell commands
// +build darwin linux

package provider_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func CheckContents(p string, ec string) resource.TestCheckFunc {
	return func(tfs *terraform.State) (err error) {
		fh, err := os.Open(p)
		if err != nil {
			return
		}
		defer fh.Close()
		bc, err := ioutil.ReadAll(fh)
		c := string(bc)
		if c != ec {
			err = errors.New("wrong contents")
		}
		return
	}
}

func TestAccResourceExternal(t *testing.T) {
	// Since this function exits before the tests necessarily run, we just leave
	// the temporary directory undeleted
	td, err := ioutil.TempDir("", "resource_external_test")
	if err != nil {
		t.Skip(err.Error())
	}
	tmplS := struct{ TempDir string }{TempDir: td}
	ts := map[string]resource.TestCase{
		"fail": {
			IsUnitTest:        true,
			ProviderFactories: ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: ReadConfig(
						t,
						filepath.Join("external", "fail", "resource-fail.hcl"),
						tmplS,
					),
					ExpectError: regexp.MustCompile("exit status"),
				},
			},
		},
		"file": {
			IsUnitTest:        true,
			ProviderFactories: ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: fmt.Sprintf(`
					provider packernix{}
					resource packernix_external file {
						path = "./testdata/external/file"
						options = [ "%s/file.txt", "file content" ]
					}
					`, td),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr(
							"packernix_external.file",
							"state",
							regexp.MustCompile(regexp.QuoteMeta(
								fmt.Sprintf(
									"%s/file.txt",
									td,
								),
							)),
						),
						CheckContents(
							filepath.Join(td, "file.txt"),
							"file content",
						),
					),
				},
				{
					Config: fmt.Sprintf(`
					provider packernix{}
					resource packernix_external file {
						path = "./testdata/external/file"
						options = [ "%s/file.txt", "new content" ]
					}
					`, td),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr(
							"packernix_external.file",
							"state",
							regexp.MustCompile(regexp.QuoteMeta(
								fmt.Sprintf(
									"%s/file.txt",
									td,
								),
							)),
						),
						CheckContents(
							filepath.Join(td, "file.txt"),
							"new content",
						),
					),
				},
				{
					ResourceName: "packernix_external.file",
					ImportState:  true,
				},
				{
					Config: "provider packernix{}",
					Check: func(tfs *terraform.State) error {
						_, err := os.Stat(
							filepath.Join(td, "file.txt"),
						)
						if !os.IsNotExist(err) {
							return errors.New("file still exists")
						}
						return nil
					},
				},
			},
		},
		"hash": {
			IsUnitTest:        true,
			ProviderFactories: ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: fmt.Sprintf(`
					provider packernix{
						working_dir = "./testdata/external/namedfile"
					}
					resource packernix_external hash {
						path = "."
						options = [ "%s/hash.txt", "hash test file content" ]
					}
					`, td),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr(
							"packernix_external.hash",
							"state",
							regexp.MustCompile(regexp.QuoteMeta(
								fmt.Sprintf(
									"%s/hash.txt:hash test file content",
									td,
								),
							)),
						),
						CheckContents(
							filepath.Join(td, "hash.txt"),
							"hash test file content",
						),
					),
				},
				{
					Config: fmt.Sprintf(`
					provider packernix{
						working_dir = "./testdata/external/namedfile_extra_parent/testdata/external/namedfile"
					}
					resource packernix_external hash {
						path = "."
						options = [ "%s/hash.txt", "hash test file content" ]
					}
					`, td),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr(
							"packernix_external.hash",
							"state",
							regexp.MustCompile(regexp.QuoteMeta(
								fmt.Sprintf(
									"%s/hash.txt:hash test file content",
									td,
								),
							)),
						),
						CheckContents(
							filepath.Join(td, "hash.txt"),
							"hash test file content",
						),
					),
				},
				{
					ResourceName: "packernix_external.hash",
					ImportState:  true,
				},
				{
					Config: fmt.Sprintf(`
					provider packernix{
						working_dir = "./testdata/external/namedfile_bad_hash"
					}
					resource packernix_external hash {
						path = "."
						options = [ "%s/hash.txt", "hash test file content" ]
					}
`, td),
					ExpectError: regexp.MustCompile("hash mismatch"),
				},
				// we have to manually destroy resource since . does not have
				// the same hash as "./testdata/external/namedfile"
				{
					Config: `
					provider packernix{
						working_dir = "./testdata/external/namedfile"
					}
					`,
					Check: func(tfs *terraform.State) error {
						_, err := os.Stat(
							filepath.Join(td, "hash.txt"),
						)
						if !os.IsNotExist(err) {
							return errors.New("file still exists")
						}
						return nil
					},
				},
			},
		},
		"named file": {
			IsUnitTest:        true,
			ProviderFactories: ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: fmt.Sprintf(`
					provider packernix{}
					resource packernix_external namedfile {
						path = "./testdata/external/namedfile"
						options = [ "%s/namedfile.txt", "named file content" ]
					}
					`, td),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr(
							"packernix_external.namedfile",
							"state",
							regexp.MustCompile(regexp.QuoteMeta(
								fmt.Sprintf(
									"%s/namedfile.txt:named file content",
									td,
								),
							)),
						),
						CheckContents(
							filepath.Join(td, "namedfile.txt"),
							"named file content",
						),
					),
				},
				{
					Config: fmt.Sprintf(`
					provider packernix{}
					resource packernix_external namedfile {
						path = "./testdata/external/namedfile"
						options = [ "%s/namedfile.txt", "named new content" ]
					}
					`, td),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr(
							"packernix_external.namedfile",
							"state",
							regexp.MustCompile(regexp.QuoteMeta(
								fmt.Sprintf(
									"%s/namedfile.txt:named new content",
									td,
								),
							)),
						),
						CheckContents(
							filepath.Join(td, "namedfile.txt"),
							"named new content",
						),
					),
				},
				{
					ResourceName: "packernix_external.namedfile",
					ImportState:  true,
				},
				{
					Config: "provider packernix{}",
					Check: func(tfs *terraform.State) error {
						_, err := os.Stat(
							filepath.Join(td, "namedfile.txt"),
						)
						if !os.IsNotExist(err) {
							return errors.New("file still exists")
						}
						return nil
					},
				},
			},
		},
		"noop": {
			IsUnitTest:        true,
			ProviderFactories: ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: ReadConfig(
						t,
						filepath.Join("external", "noop", "resource-noop.hcl"),
						tmplS,
					),
					ExpectError: regexp.MustCompile("empty string"),
				},
			},
		},
		"working_dir swapping": {
			IsUnitTest:        true,
			ProviderFactories: ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: fmt.Sprintf(`
					provider packernix{
						working_dir = "./testdata/external/namedfile_extra_parent"
					}
					resource packernix_external wdswap {
						path = "./testdata/external/namedfile"
						options = [ "%s/wdswap.txt", "step1" ]
					}
					`, td),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr(
							"packernix_external.wdswap",
							"state",
							regexp.MustCompile(regexp.QuoteMeta(
								fmt.Sprintf(
									"%s/wdswap.txt:step1",
									td,
								),
							)),
						),
						CheckContents(
							filepath.Join(td, "wdswap.txt"),
							"step1",
						),
					),
				},
				{
					Config: fmt.Sprintf(`
					provider packernix{
						working_dir = "."
					}
					resource packernix_external wdswap {
						path = "./testdata/external/namedfile"
						options = [ "%s/wdswap.txt", "step2" ]
					}
					`, td),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr(
							"packernix_external.wdswap",
							"state",
							regexp.MustCompile(regexp.QuoteMeta(
								fmt.Sprintf(
									"%s/wdswap.txt:step2",
									td,
								),
							)),
						),
						CheckContents(
							filepath.Join(td, "wdswap.txt"),
							"step2",
						),
					),
				},
				{
					Config: fmt.Sprintf(`
					provider packernix{}
					resource packernix_external wdswap {
						working_dir = "./testdata"
						path = "./external/namedfile"
						options = [ "%s/wdswap.txt", "step3" ]
					}
					`, td),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr(
							"packernix_external.wdswap",
							"state",
							regexp.MustCompile(regexp.QuoteMeta(
								fmt.Sprintf(
									"%s/wdswap.txt:step3",
									td,
								),
							)),
						),
						CheckContents(
							filepath.Join(td, "wdswap.txt"),
							"step3",
						),
					),
				},
				{
					Config: fmt.Sprintf(`
					provider packernix{}
					resource packernix_external wdswap {
						working_dir = "testdata/external"
						path = "./namedfile"
						options = [ "%s/wdswap.txt", "step4" ]
					}
					`, td),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr(
							"packernix_external.wdswap",
							"state",
							regexp.MustCompile(regexp.QuoteMeta(
								fmt.Sprintf(
									"%s/wdswap.txt:step4",
									td,
								),
							)),
						),
						CheckContents(
							filepath.Join(td, "wdswap.txt"),
							"step4",
						),
					),
				},
				{
					Config: fmt.Sprintf(`
					provider packernix{
						working_dir = "./testdata/external/namedfile_extra_parent"
					}
					resource packernix_external wdswap {
						path = "./testdata/external/namedfile"
						options = [ "%s/wdswap.txt", "step5" ]
					}
					`, td),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr(
							"packernix_external.wdswap",
							"state",
							regexp.MustCompile(regexp.QuoteMeta(
								fmt.Sprintf(
									"%s/wdswap.txt:step5",
									td,
								),
							)),
						),
						CheckContents(
							filepath.Join(td, "wdswap.txt"),
							"step5",
						),
					),
				},
				{
					Config: fmt.Sprintf(`
					provider packernix{}
					resource packernix_external wdswap {
						working_dir = "./testdata/external"
						path = "./namedfile"
						options = [ "%s/wdswap.txt", "step6" ]
					}
					`, td),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr(
							"packernix_external.wdswap",
							"state",
							regexp.MustCompile(regexp.QuoteMeta(
								fmt.Sprintf(
									"%s/wdswap.txt:step6",
									td,
								),
							)),
						),
						CheckContents(
							filepath.Join(td, "wdswap.txt"),
							"step6",
						),
					),
				},
				{
					ResourceName: "packernix_external.wdswap",
					ImportState:  true,
				},
			},
		},
		"program swapping": {
			IsUnitTest:        true,
			ProviderFactories: ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: fmt.Sprintf(`
					provider packernix{}
					resource packernix_external pswap {
						path = "./testdata/external/file"
						options = [ "%s/programswap.txt", "step1" ]
					}
					`, td),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr(
							"packernix_external.pswap",
							"state",
							regexp.MustCompile(regexp.QuoteMeta(
								fmt.Sprintf(
									"%s/programswap.txt",
									td,
								),
							)),
						),
						CheckContents(
							filepath.Join(td, "programswap.txt"),
							"step1",
						),
					),
				},
				{
					Config: fmt.Sprintf(`
					provider packernix{}
					resource packernix_external pswap {
						path = "./testdata/external/file"
						options = [ "%s/programswap.txt", "step2" ]
					}
					`, td),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr(
							"packernix_external.pswap",
							"state",
							regexp.MustCompile(regexp.QuoteMeta(
								fmt.Sprintf(
									"%s/programswap.txt",
									td,
								),
							)),
						),
						CheckContents(
							filepath.Join(td, "programswap.txt"),
							"step2",
						),
					),
				},
				{
					Config: fmt.Sprintf(`
					provider packernix{}
					resource packernix_external pswap {
						path = "./testdata/external/namedfile"
						options = [ "%s/programswap.txt", "step3" ]
					}
					`, td),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr(
							"packernix_external.pswap",
							"state",
							regexp.MustCompile(regexp.QuoteMeta(
								fmt.Sprintf(
									"%s/programswap.txt:step3",
									td,
								),
							)),
						),
						CheckContents(
							filepath.Join(td, "programswap.txt"),
							"step3",
						),
					),
				},
				{
					Config: fmt.Sprintf(`
					provider packernix{}
					resource packernix_external pswap {
						path = "./testdata/external/file"
						options = [ "%s/programswap.txt", "step4" ]
					}
					`, td),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr(
							"packernix_external.pswap",
							"state",
							regexp.MustCompile(regexp.QuoteMeta(
								fmt.Sprintf(
									"%s/programswap.txt",
									td,
								),
							)),
						),
						CheckContents(
							filepath.Join(td, "programswap.txt"),
							"step4",
						),
					),
				},
				{
					Config: fmt.Sprintf(`
					provider packernix{}
					resource packernix_external pswap {
						path = "./testdata/external/namedfile"
						options = [ "%s/programswap.txt", "step5" ]
					}
					`, td),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr(
							"packernix_external.pswap",
							"state",
							regexp.MustCompile(regexp.QuoteMeta(
								fmt.Sprintf(
									"%s/programswap.txt:step5",
									td,
								),
							)),
						),
						CheckContents(
							filepath.Join(td, "programswap.txt"),
							"step5",
						),
					),
				},
				{
					Config: fmt.Sprintf(`
					provider packernix{}
					resource packernix_external pswap {
						path = "./testdata/external/namedfile"
						options = [ "%s/programswap.txt", "laststep" ]
					}
					`, td),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr(
							"packernix_external.pswap",
							"state",
							regexp.MustCompile(regexp.QuoteMeta(
								fmt.Sprintf(
									"%s/programswap.txt:laststep",
									td,
								),
							)),
						),
						CheckContents(
							filepath.Join(td, "programswap.txt"),
							"laststep",
						),
					),
				},
				{
					Config: fmt.Sprintf(`
					provider packernix{}
					resource packernix_external pswap {
						path = "./testdata/external/namedfile_bad_hash"
						options = [ "%s/programswap.txt", "laststep" ]
					}
					`, td),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr(
							"packernix_external.pswap",
							"state",
							regexp.MustCompile(regexp.QuoteMeta(
								fmt.Sprintf(
									"%s/programswap.txt:laststep",
									td,
								),
							)),
						),
						CheckContents(
							filepath.Join(td, "programswap.txt"),
							"laststep",
						),
					),
				},
				{
					Config: fmt.Sprintf(`
					provider packernix{}
					resource packernix_external pswap {
						path = "./testdata/external/file"
						options = [ "%s/programswap.txt", "laststep" ]
					}
					`, td),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr(
							"packernix_external.pswap",
							"state",
							regexp.MustCompile(regexp.QuoteMeta(
								fmt.Sprintf(
									"%s/programswap.txt",
									td,
								),
							)),
						),
						CheckContents(
							filepath.Join(td, "programswap.txt"),
							"laststep",
						),
					),
				},
				{
					ResourceName: "packernix_external.pswap",
					ImportState:  true,
				},
			},
		},
	}
	for k, tt := range ts {
		t.Run(k, func(t *testing.T) {
			resource.ParallelTest(t, tt)
		})
	}
}
