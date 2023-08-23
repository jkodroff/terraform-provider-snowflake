package generator

import "text/template"

var PackageTemplate, _ = template.New("packageTemplate").Parse(`
package {{.}}
`)

var InterfaceTemplate, _ = template.New("interfaceTemplate").Parse(`
import "context"

type {{.Name}} interface {
	{{- range .Operations}}
		{{.Name}}(ctx context.Context, request *{{.DtoName}}) error
	{{- end}}
}
`)

var OptionsTemplate, _ = template.New("optionsTemplate").Parse(`
// {{.OptsName}} is based on {{.Doc}}.
type {{.OptsName}} struct {
	{{- range .Fields}}
			{{.Name}} {{.Kind}} {{.TagsPrintable}}
	{{- end}}
}
`)

var StructTemplate, _ = template.New("structTemplate").Parse(`
type {{.KindNoPtr}} struct {
	{{- range .Fields}}
			{{.Name}} {{.Kind}} {{.TagsPrintable}}
	{{- end}}
}
`)

var DtoTemplate, _ = template.New("dtoTemplate").Parse(`
{{define "DTO_STRUCT"}}
type {{.DtoName}} struct {
	{{- range .Fields}}
		{{- if .ShouldBeInDto}}
		{{.Name}} {{.DtoKind}} {{if .Required}}// required{{end}}
		{{- end}}
	{{- end}}
}
{{end}}

//go:generate go run ../../dto-builder-generator/main.go

var (
	{{- range .Operations}}
	_ optionsProvider[{{.OptsName}}] = new({{.DtoName}})
	{{- end}}
)

{{- range .Operations}}
	{{template "DTO_STRUCT" .}}
	{{- range .Fields}}
		{{if .IsStruct}}
		{{template "DTO_STRUCT" .}}
		{{end}}
	{{- end}}
{{- end}}
`)

var ImplementationTemplate, _ = template.New("implementationTemplate").Parse(`
{{define "MAPPING" -}}
	&{{.KindNoPtr}}{
		{{- range .Fields}}
			{{- if .ShouldBeInDto}}
			{{if .IsStruct}}{{else}}{{.Name}}: r.{{.Name}},{{end}}
			{{- end}}
		{{- end}}
	}
	{{range .Fields}}
		{{if .ShouldBeInDto}}
			{{if .IsStruct}}
				if r.{{.Name}} != nil {
					opts.{{.Name}} = {{template "MAPPING" .}}
				}
			{{end}}
		{{end}}
	{{end}}
{{end}}
import "context"

{{$impl := .NameLowerCased}}
var _ {{.Name}} = (*{{$impl}})(nil)

type {{$impl}} struct {
	client *Client
}
{{range .Operations}}
func (v *{{$impl}}) {{.Name}}(ctx context.Context, request *{{.DtoName}}) error {
	opts := request.toOpts()
	return validateAndExec(v.client, ctx, opts)
}
{{end}}

{{range .Operations}}
func (r *{{.DtoName}}) toOpts() *{{.OptsName}} {
	opts := {{template "MAPPING" .}}
	return opts
}
{{end}}
`)

var TestFuncTemplate, _ = template.New("testFuncTemplate").Parse(`
{{define "VALIDATION_TEST"}}
	{{- range .Validations}}
	{{.TodoComment}}
	{{- end}}
{{end}}

{{define "VALIDATIONS"}}
	{{template "VALIDATION_TEST" .}}
	{{- range .AdditionalValidations}}
		{{template "VALIDATION_TEST" .}}
	{{- end}}
{{end}}

import "testing"

{{- range .Operations}}
func Test{{.ObjectInterface.Name}}_{{.Name}}(t *testing.T) {
	id := random{{.ObjectInterface.IdentifierKind}}(t)

	defaultOpts := func() *{{.OptsName}} {
		return &{{.OptsName}}{
			name: id,
		}
	}

	// TODO: fill me
	{{template "VALIDATIONS" .}}
}
{{end}}
`)

var ValidationsImplTemplate, _ = template.New("validationsImplTemplate").Parse(`
{{define "VALIDATIONS"}}
	{{- range .Validations}}
	if {{.Condition}} {
		errs = append(errs, {{.Error}})
	}
	{{- end}}
	{{- range .AdditionalValidations}}
	if {{.NameLowerCased}} := {{.Name}}; valueSet({{.NameLowerCased}}) {
		{{template "VALIDATIONS" .}}
	}
	{{- end}}
{{end}}

var (
{{- range .Operations}}
	_ validatable = new({{.OptsName}})
{{- end}}
)
{{range .Operations}}
func (opts *{{.OptsName}}) validate() error {
	if opts == nil {
		return errors.Join(errNilOptions)
	}
	var errs []error
	{{template "VALIDATIONS" .}}
	return errors.Join(errs...)
}
{{end}}
`)

var IntegrationTestsTemplate, _ = template.New("integrationTestsTemplate").Parse(`
import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInt_{{.Name}}(t *testing.T) {
	// TODO: fill me
}
`)
