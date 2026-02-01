{{- $D := or .Web .MQ .Job -}}
package conversion

import (
	"github.com/onexstack/onexstack/pkg/core"

	"{{.M.ModuleName}}/internal/{{$D.Name}}/model"
	{{$D.APIImportPath}}
)

// {{$D.R.MapModelToAPIFunc}} converts a {{$D.R.GORMModel}} object from the internal model
// to a {{$D.R.SingularName}} object in the {{.M.APIAlias}} API format.
func {{$D.R.MapModelToAPIFunc}}({{$D.R.SingularLowerFirst}}M *model.{{$D.R.GORMModel}}) *{{.M.APIAlias}}.{{$D.R.SingularName}} {
	var {{$D.R.SingularLowerFirst}} {{.M.APIAlias}}.{{$D.R.SingularName}}
	_ = core.CopyWithConverters(&{{$D.R.SingularLowerFirst}}, {{$D.R.SingularLowerFirst}}M)
	return &{{$D.R.SingularLowerFirst}}
}

// {{$D.R.MapAPIToModelFunc}} converts a {{$D.R.SingularName}} object from the {{.M.APIAlias}} API format
// to a {{$D.R.GORMModel}} object in the internal model.
func {{$D.R.MapAPIToModelFunc}}({{$D.R.SingularLowerFirst}} *{{.M.APIAlias}}.{{$D.R.SingularName}}) *model.{{$D.R.GORMModel}} {
	var {{$D.R.SingularLowerFirst}}M model.{{$D.R.GORMModel}}
	_ = core.CopyWithConverters(&{{$D.R.SingularLowerFirst}}M, {{$D.R.SingularLowerFirst}})
	return &{{$D.R.SingularLowerFirst}}M
}

