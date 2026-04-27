{{- $D := or .Web .MQ .Job -}}
package errno

import (
	"net/http"

	"github.com/onexstack/onexstack/pkg/errorsx"
)

// Err{{$D.R.SingularName}}NotFound indicates that the specified {{$D.R.SingularLower}} was not found.
var Err{{$D.R.SingularName}}NotFound = errorsx.New(http.StatusNotFound, "NotFound.{{$D.R.SingularName}}NotFound", "The requested {{$D.R.SingularLower}} was not found.")

// Err{{$D.R.SingularName}}CreateFailed indicates that the {{$D.R.SingularLower}} creation operation failed.
var Err{{$D.R.SingularName}}CreateFailed = errorsx.New(http.StatusInternalServerError, "InternalError.{{$D.R.SingularName}}CreateFailed", "failed to create the {{$D.R.SingularLower}}.")

// Err{{$D.R.SingularName}}UpdateFailed indicates that the {{$D.R.SingularLower}} update operation failed.
var Err{{$D.R.SingularName}}UpdateFailed = errorsx.New(http.StatusInternalServerError, "InternalError.{{$D.R.SingularName}}UpdateFailed", "failed to update the {{$D.R.SingularLower}}.")

// Err{{$D.R.SingularName}}DeleteFailed indicates that the {{$D.R.SingularLower}} deletion operation failed.
var Err{{$D.R.SingularName}}DeleteFailed = errorsx.New(http.StatusInternalServerError, "InternalError.{{$D.R.SingularName}}DeleteFailed", "failed to delete the {{$D.R.SingularLower}}.")

// Err{{$D.R.SingularName}}GetFailed indicates that retrieving the specified {{$D.R.SingularLower}} failed.
var Err{{$D.R.SingularName}}GetFailed = errorsx.New(http.StatusInternalServerError, "InternalError.{{$D.R.SingularName}}GetFailed", "failed to retrieve the {{$D.R.SingularLower}} details.")

// Err{{$D.R.SingularName}}ListFailed indicates that listing {{$D.R.PluralLower}} failed.
var Err{{$D.R.SingularName}}ListFailed = errorsx.New(http.StatusInternalServerError, "InternalError.{{$D.R.SingularName}}ListFailed", "failed to list {{$D.R.PluralLower}}.")
