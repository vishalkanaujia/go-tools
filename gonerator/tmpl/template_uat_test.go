// Copyright 2017 Corey Scott http://www.sage42.org/
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

package tmpl

import (
	"bytes"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUAT_template(t *testing.T) {
	scenarios := []struct {
		desc       string
		inTemplate string
		inSrc      string
		expected   string
		expectErr  bool
	}{
		{
			desc:       "fields list using range",
			inTemplate: `row.Scan({{$len := len .Fields }}{{range $index, $value := .Fields}}&in.{{$value.Name}}{{isNotLast $len $index ", "}}{{end}})`,
			inSrc: `package test

type myType struct {
	ID      int64
	Name    string
	Balance float64
}
`,
			expected:  `row.Scan(&in.ID, &in.Name, &in.Balance)`,
			expectErr: false,
		},
		{
			desc:       "fields list using range - extended",
			inTemplate: `{{ $typeName := .TypeName }}{{ $len := len .Methods }}{{ range $index, $value := .Methods }}func (impl {{ $typeName }}Impl) {{ $value.Name }}({{ $plen := len $value.Params }}{{ range $pindex, $pvalue := $value.Params }}{{ $pplen := len $pvalue.Names }}{{ range $ppindex, $ppname := $pvalue.Names }}{{ $ppname }}{{ isNotLast $pplen $ppindex ", " }}{{ end }} {{ $pvalue.Type }}{{ isNotLast $plen $pindex ", " }}{{ end }}) {{ $rlen := len .Results }}({{ range $rindex, $rvalue := .Results }}{{ $rvalue.Type }}{{ isNotLast $rlen $rindex ", " }}{{ end }}) {}{{ end }}`,
			inSrc: `package test

type myType interface {
	LoadByID(ctx context.Context, id int64) tType
}
`,
			expected:  `func (impl myTypeImpl) LoadByID(ctx context.Context, id int64) (tType) {}`,
			expectErr: false,
		},
		{
			desc:       "fieldsList",
			inTemplate: `row.Scan({{ fieldsList .Fields "&in.{{$field.Name}}" }})`,
			inSrc: `package test

type myType struct {
	ID      int64
	Name    string
	Balance float64
}
`,
			expected:  `row.Scan(&in.ID, &in.Name, &in.Balance)`,
			expectErr: false,
		},
		{
			desc:       "fieldsListWithTag",
			inTemplate: `row.Scan({{ fieldsListWithTag .Fields "&in.{{$field.Name}}" "sql-col" }})`,
			inSrc: `package test

type myType struct {
	ID      int64	` + "`" + `sql-col:"id"` + "`" + `
	Name    string
	Balance float64
}
`,
			expected:  `row.Scan(&in.ID)`,
			expectErr: false,
		},
		{
			desc:       "fieldsListWithTagValue",
			inTemplate: `row.Scan({{ fieldsListWithTagValue .Fields "&in.{{$field.Name}}" "sql-key" "false" }})`,
			inSrc: `package test

type myType struct {
	ID      int64	` + "`" + `sql-key:"true"` + "`" + `
	Name    string  ` + "`" + `sql-key:"false"` + "`" + `
	Balance float64 ` + "`" + `sql-key:"false"` + "`" + `
}
`,
			expected:  `row.Scan(&in.Name, &in.Balance)`,
			expectErr: false,
		},
	}

	for _, s := range scenarios {
		scenario := s
		t.Run(scenario.desc, func(t *testing.T) {
			typeName := "myType"
			vars := TemplateData{
				TypeName: typeName,
				Fields:   GetFields(getASTFromSrc(scenario.inSrc), typeName),
				Methods:  GetMethods(getASTFromSrc(scenario.inSrc), typeName),
			}

			parsedTemplate, err := getTemplate().Parse(scenario.inTemplate)
			if err != nil {
				log.Fatal(err)
			}

			buffer := &bytes.Buffer{}
			resultErr := parsedTemplate.Execute(buffer, vars)
			result := buffer.String()

			require.Equal(t, scenario.expectErr, resultErr != nil, "expected error", resultErr)
			assert.Equal(t, scenario.expected, result, "expected result")
		})
	}
}

func TestUAT_noop(t *testing.T) {
	scenarios := []struct {
		desc      string
		inSrc     string
		expected  string
		expectErr bool
	}{
		{
			desc: "simple",
			inSrc: `package test

type myInterface interface {
	ID() int64
	Name() string 
	Balance() float64
	Login(username, password string) (Session, error)
	Record(message string, args ...interface{})
}
`,
			expected: `
package test

type noopMyInterface struct {}

func (*noopMyInterface) ID() (_ int64) { return }

func (*noopMyInterface) Name() (_ string) { return }

func (*noopMyInterface) Balance() (_ float64) { return }

func (*noopMyInterface) Login(_, _ string) (_ Session, _ error) { return }

func (*noopMyInterface) Record(_ string, _ ...interface{}) {}

`,
			expectErr: false,
		},
	}

	for _, s := range scenarios {
		scenario := s
		t.Run(scenario.desc, func(t *testing.T) {
			typeName := "myInterface"

			vars := TemplateData{
				TypeName:    typeName,
				PackageName: "test",
				Fields:      GetFields(getASTFromSrc(scenario.inSrc), typeName),
				Methods:     GetMethods(getASTFromSrc(scenario.inSrc), typeName),
			}

			parsedTemplate, err := getTemplate().Parse(NoopTemplate)
			if err != nil {
				log.Fatal(err)
			}

			buffer := &bytes.Buffer{}
			resultErr := parsedTemplate.Execute(buffer, vars)
			result := buffer.String()

			require.Equal(t, scenario.expectErr, resultErr != nil, "expected error", resultErr)
			assert.Equal(t, scenario.expected, result, "expected result")
		})
	}
}
