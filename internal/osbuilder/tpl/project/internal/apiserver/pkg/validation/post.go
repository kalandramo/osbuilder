{{- $D := or .Web .MQ .Job -}}
package validation

import (
	"context"

	genericvalidation "github.com/onexstack/onexstack/pkg/validation"

	{{$D.APIImportPath}}
)

// Validate{{$D.R.SingularName}}Rules returns a set of validation rules for {{$D.R.SingularLower}}-related requests.
func (v *Validator) Validate{{$D.R.SingularName}}Rules() genericvalidation.Rules {
	return genericvalidation.Rules{}
}

// ValidateCreate{{$D.R.SingularName}}Request validates the fields of a Create{{$D.R.SingularName}}Request.
func (v *Validator) ValidateCreate{{$D.R.SingularName}}Request(ctx context.Context, rq *{{.M.APIAlias}}.Create{{$D.R.SingularName}}Request) error {
	return genericvalidation.ValidateAllFields(rq, v.Validate{{$D.R.SingularName}}Rules())
}

// ValidateUpdate{{$D.R.SingularName}}Request validates the fields of an Update{{$D.R.SingularName}}Request.
func (v *Validator) ValidateUpdate{{$D.R.SingularName}}Request(ctx context.Context, rq *{{.M.APIAlias}}.Update{{$D.R.SingularName}}Request) error {
	return genericvalidation.ValidateAllFields(rq, v.Validate{{$D.R.SingularName}}Rules())
}

// ValidateDelete{{$D.R.SingularName}}Request validates the fields of a Delete{{$D.R.SingularName}}Request.
func (v *Validator) ValidateDelete{{$D.R.SingularName}}Request(ctx context.Context, rq *{{.M.APIAlias}}.Delete{{$D.R.SingularName}}Request) error {
	return genericvalidation.ValidateAllFields(rq, v.Validate{{$D.R.SingularName}}Rules())
}

// ValidateDelete{{$D.R.PluralName}}Request validates the fields of a Delete{{$D.R.PluralName}}Request.
func (v *Validator) ValidateDelete{{$D.R.PluralName}}Request(ctx context.Context, rq *v1.Delete{{$D.R.PluralName}}Request) error {
    return genericvalidation.ValidateAllFields(rq, v.Validate{{$D.R.SingularName}}Rules())
}

// ValidateGet{{$D.R.SingularName}}Request validates the fields of a Get{{$D.R.SingularName}}Request.
func (v *Validator) ValidateGet{{$D.R.SingularName}}Request(ctx context.Context, rq *{{.M.APIAlias}}.Get{{$D.R.SingularName}}Request) error {
	return genericvalidation.ValidateAllFields(rq, v.Validate{{$D.R.SingularName}}Rules())
}

// ValidateList{{$D.R.SingularName}}Request validates the fields of a List{{$D.R.SingularName}}Request, focusing on selected fields ("Offset" and "Limit").
func (v *Validator) ValidateList{{$D.R.SingularName}}Request(ctx context.Context, rq *{{.M.APIAlias}}.List{{$D.R.SingularName}}Request) error {
	return genericvalidation.ValidateSelectedFields(rq, v.Validate{{$D.R.SingularName}}Rules(), "Offset", "Limit")
}
