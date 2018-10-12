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
	"github.com/palantir/godel/framework/pluginapitester"
	"github.com/palantir/godel/pkg/products/v2/products"
	"github.com/palantir/okgo/okgotester"
	"github.com/stretchr/testify/require"
)

const (
	okgoPluginLocator  = "com.palantir.okgo:check-plugin:1.0.0"
	okgoPluginResolver = "https://palantir.bintray.com/releases/{{GroupPath}}/{{Product}}/{{Version}}/{{Product}}-{{Version}}-{{OS}}-{{Arch}}.tgz"
)

func TestCheck(t *testing.T) {
	const godelYML = `exclude:
  names:
    - "\\..+"
    - "vendor"
  paths:
    - "godel"
`

	assetPath, err := products.Bin("importalias-asset")
	require.NoError(t, err)

	configFiles := map[string]string{
		"godel/config/godel.yml":        godelYML,
		"godel/config/check-plugin.yml": "",
	}

	pluginProvider, err := pluginapitester.NewPluginProviderFromLocator(okgoPluginLocator, okgoPluginResolver)
	require.NoError(t, err)

	okgotester.RunAssetCheckTest(t,
		pluginProvider,
		pluginapitester.NewAssetProvider(assetPath),
		"importalias",
		"",
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
Check(s) produced output: [importalias]
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
Check(s) produced output: [importalias]
`,
			},
		},
	)
}

func TestUpgradeConfig(t *testing.T) {
	pluginProvider, err := pluginapitester.NewPluginProviderFromLocator(okgoPluginLocator, okgoPluginResolver)
	require.NoError(t, err)

	assetPath, err := products.Bin("importalias-asset")
	require.NoError(t, err)
	assetProvider := pluginapitester.NewAssetProvider(assetPath)

	pluginapitester.RunUpgradeConfigTest(t,
		pluginProvider,
		[]pluginapitester.AssetProvider{assetProvider},
		[]pluginapitester.UpgradeConfigTestCase{
			{
				Name: `legacy configuration with empty "args" field is updated`,
				ConfigFiles: map[string]string{
					"godel/config/check.yml": `
checks:
  importalias:
    filters:
      - value: "should have comment or be unexported"
      - type: name
        value: ".*.pb.go"
`,
				},
				Legacy: true,
				WantOutput: `Upgraded configuration for check-plugin.yml
`,
				WantFiles: map[string]string{
					"godel/config/check-plugin.yml": `checks:
  importalias:
    filters:
    - value: should have comment or be unexported
    exclude:
      names:
      - .*.pb.go
`,
				},
			},
			{
				Name: `legacy configuration with non-empty "args" field fails`,
				ConfigFiles: map[string]string{
					"godel/config/check.yml": `
checks:
  importalias:
    args:
      - "-foo"
`,
				},
				Legacy:    true,
				WantError: true,
				WantOutput: `Failed to upgrade configuration:
	godel/config/check-plugin.yml: failed to upgrade configuration: failed to upgrade check "importalias" legacy configuration: failed to upgrade asset configuration: importalias-asset does not support legacy configuration with a non-empty "args" field
`,
				WantFiles: map[string]string{
					"godel/config/check.yml": `
checks:
  importalias:
    args:
      - "-foo"
`,
				},
			},
			{
				Name: `empty v0 config works`,
				ConfigFiles: map[string]string{
					"godel/config/check-plugin.yml": `
checks:
  importalias:
    skip: true
    # comment preserved
    config:
`,
				},
				WantOutput: ``,
				WantFiles: map[string]string{
					"godel/config/check-plugin.yml": `
checks:
  importalias:
    skip: true
    # comment preserved
    config:
`,
				},
			},
			{
				Name: `non-empty v0 config does not work`,
				ConfigFiles: map[string]string{
					"godel/config/check-plugin.yml": `
checks:
  importalias:
    config:
      # comment
      key: value
`,
				},
				WantError: true,
				WantOutput: `Failed to upgrade configuration:
	godel/config/check-plugin.yml: failed to upgrade check "importalias" configuration: failed to upgrade asset configuration: importalias-asset does not currently support configuration
`,
				WantFiles: map[string]string{
					"godel/config/check-plugin.yml": `
checks:
  importalias:
    config:
      # comment
      key: value
`,
				},
			},
		},
	)
}
