package handler

import (
	"github.com/google/wire"

    "{{.M.ModuleName}}/internal/{{.Web.Name}}/biz"
	{{.Web.APIImportPath}}
)

// ProviderSet contains providers for creating instances of the biz struct.
var ProviderSet = wire.NewSet(NewHandler, wire.Bind(new({{.M.APIVersion}}.{{.Web.GRPCServiceName}}Server), new(*Handler)))

// Handler implements a gRPC service.
type Handler struct {
	{{.M.APIAlias}}.Unimplemented{{.Web.GRPCServiceName}}Server

	biz biz.IBiz
}

// Ensure that Handler implements the {{.M.APIVersion}}.{{.Web.GRPCServiceName}}Server interface.
var _ {{.M.APIVersion}}.{{.Web.GRPCServiceName}}Server = (*Handler)(nil)

// NewHandler creates a new instance of *Handler.
func NewHandler(biz biz.IBiz) *Handler {
	return &Handler{biz: biz}
}
