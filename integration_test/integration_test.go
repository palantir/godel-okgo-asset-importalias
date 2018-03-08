// Copyright 2016 Palantir Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package integration_test

import (
	"testing"

	"github.com/nmiyake/pkg/gofiles"
	"github.com/palantir/godel/pkg/products"
	"github.com/palantir/okgo/okgotester"
	"github.com/stretchr/testify/require"
)

const (
	okgoPluginLocator  = "com.palantir.okgo:okgo-plugin:0.3.0"
	okgoPluginResolver = "https://palantir.bintray.com/releases/{{GroupPath}}/{{Product}}/{{Version}}/{{Product}}-{{Version}}-{{OS}}-{{Arch}}.tgz"

	godelYML = `exclude:
  names:
    - "\\..+"
    - "vendor"
  paths:
    - "godel"
`
)

func TestImportAlias(t *testing.T) {
	assetPath, err := products.Bin("importalias-asset")
	require.NoError(t, err)

	configFiles := map[string]string{
		"godel/config/godel.yml": godelYML,
		"godel/config/check.yml": "",
	}

	okgotester.RunAssetCheckTest(t,
		okgoPluginLocator, okgoPluginResolver,
		assetPath, "importalias",
		[]okgotester.AssetTestCase{
			{
				Name: "importalias used inconsistently",
				Specs: []gofiles.GoFileSpec{
					{
						RelPath: "foo.go",
						Src:     `package main; import foo "fmt"; func main(){ foo.Println() }`,
					},
					{
						RelPath: "bar/bar.go",
						Src:     `package bar; import bar "fmt"; func Bar(){ bar.Println() }`,
					},
					{
						RelPath: "baz/baz.go",
						Src:     `package baz; import foo "fmt"; func Baz(){ bar.Println() }`,
					},
				},
				ConfigFiles: configFiles,
				WantError:   true,
				WantOutput: `Running importalias...
bar/bar.go:1:21: uses alias "bar" to import package "fmt". Use alias "foo" instead.
Finished importalias
`,
			},
			{
				Name: "importalias used inconsistently in file from inner directory",
				Specs: []gofiles.GoFileSpec{
					{
						RelPath: "foo.go",
						Src:     `package main; import foo "fmt"; func main(){ foo.Println() }`,
					},
					{
						RelPath: "bar/bar.go",
						Src:     `package bar; import bar "fmt"; func Bar(){ bar.Println() }`,
					},
					{
						RelPath: "baz/baz.go",
						Src:     `package baz; import foo "fmt"; func Baz(){ bar.Println() }`,
					},
					{
						RelPath: "inner/bar",
					},
				},
				ConfigFiles: configFiles,
				Wd:          "inner",
				WantError:   true,
				WantOutput: `Running importalias...
../bar/bar.go:1:21: uses alias "bar" to import package "fmt". Use alias "foo" instead.
Finished importalias
`,
			},
		},
	)
}
