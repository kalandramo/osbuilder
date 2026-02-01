package types

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/duke-git/lancet/v2/strutil"
	"github.com/gobuffalo/flect"

	"github.com/onexstack/osbuilder/internal/osbuilder/known"
)

// WebServer describes a web server component to generate (HTTP/gRPC/etc).
type WebServer struct {
	// BinaryName is the CLI binary name (e.g., "mb-apiserver").
	BinaryName string `yaml:"binaryName"`
	// WebFramework selects the framework (e.g., gin, grpc).
	WebFramework string `yaml:"webFramework"`
	// GRPCServiceName is the gRPC service name; default: UpperFirst(component name).
	GRPCServiceName string `yaml:"grpcServiceName,omitempty"`
	// StorageType selects backing storage (e.g., memory, mysql).
	StorageType string `yaml:"storageType"`
	// Feature flags
	WithHealthz     bool     `yaml:"withHealthz,omitempty"`
	WithUser        bool     `yaml:"withUser,omitempty"`
	WithOTel        bool     `yaml:"withOTel,omitempty"`
	WithWS          bool     `yaml:"withWS,omitempty"`
	WithPreloader   bool     `yaml:"withPreloader,omitempty"`
	ServiceRegistry string   `yaml:"serviceRegistry,omitempty"`
	Clients         []string `yaml:"clients,omitempty"`

	// Computed/derived fields (not serialized).
	ServerGen `yaml:"-"`
}

// Complete populates derived fields and sensible defaults.
func (ws *WebServer) Complete(proj *Project) *WebServer {
	ws.ServerGen = NewServerGen(proj, ws.BinaryName)

	// Default gRPC service name to UpperFirst(ComponentName), e.g., "Apiserver".
	if strings.TrimSpace(ws.GRPCServiceName) == "" {
		ws.GRPCServiceName = strutil.UpperFirst(ws.Name)
	}

	return ws
}

// HandlerDir returns the handler directory for the component.
func (ws *WebServer) HandlerDir() string {
	return filepath.Join(ws.BaseDir(), "handler")
}

// BizDir returns the business logic directory for the component.
func (ws *WebServer) BizDir() string {
	return filepath.Join(ws.BaseDir(), "biz")
}

// RESTBizFile returns the business logic directory for the specified rest resource.
func (ws *WebServer) RESTBizFile() string {
	return filepath.Join(
		ws.BizDir(),
		ws.Proj.M.APIVersion,
		ws.R.ResourcePathPrefix,
		ws.R.REST.SingularLower,
		ws.R.REST.SingularLower+".go",
	)
}

// StoreDir returns the data store directory for the component.
func (ws *WebServer) StoreDir() string {
	return filepath.Join(ws.BaseDir(), "store")
}

// RESTStoreFile returns the path to a REST store implementation for a singular resource.
func (ws *WebServer) RESTStoreFile() string {
	return filepath.Join(ws.StoreDir(), ws.R.FileName)
}

// BaseDir returns the component base directory: internal/<component>.
func (ws *WebServer) BaseDir() string {
	return filepath.Join("internal", ws.Name)
}

// ModelDir returns the model directory for the component.
func (ws *WebServer) ModelDir() string {
	return filepath.Join(ws.BaseDir(), "model")
}

// PkgDir returns the pkg directory for the component.
func (ws *WebServer) PkgDir() string {
	return filepath.Join(ws.BaseDir(), "pkg")
}

// APIDir returns the API directory for the component: pkg/api/<component>/<version>.
func (ws *WebServer) APIDir() string {
	return filepath.Join("pkg/api", ws.Name, ws.Proj.M.APIVersion)
}

// PrepareRESTMetadata constructs REST metadata for a given kind.
func (ws *WebServer) PrepareRESTMetadata(kindPath string) {
	ws.R = prepareRESTMetadata(ws.Proj.M.APIVersion, kindPath)
}

// prepareRESTMetadata constructs REST metadata for a given kind.
func prepareRESTMetadata(apiVersion string, kindPath string) *RESTGen {
	// kindPath: jobv1/cron_job
	// lastKind: cron_job
	// kind: job_cron_job
	lastKind := filepath.Base(kindPath)
	// kind := strings.ReplaceAll(kindPath, "/", "_")
	upperVer := strings.ToUpper(apiVersion)

	r := RESTGen{
		// SingularName:       strutil.UpperFirst(strutil.CamelCase(kind)),
		// SingularLowerFirst: strutil.CamelCase(kind),
		// SingularLower:      strings.ToLower(strutil.CamelCase(kind)),
		// PluralName:         flect.Pluralize(strutil.UpperFirst(strutil.CamelCase(kind))),
		// PluralLowerFirst:   flect.Pluralize(strutil.CamelCase(kind)),
		// PluralLower:        strings.ToLower(flect.Pluralize(strutil.CamelCase(kind))),
		REST: REST{
			SingularName:       strutil.UpperFirst(strutil.CamelCase(lastKind)),
			SingularLowerFirst: strutil.CamelCase(lastKind),
			SingularLower:      strings.ToLower(strutil.CamelCase(lastKind)),
			PluralName:         flect.Pluralize(strutil.UpperFirst(strutil.CamelCase(lastKind))),
			PluralLowerFirst:   flect.Pluralize(strutil.CamelCase(lastKind)),
			PluralLower:        strings.ToLower(flect.Pluralize(strutil.CamelCase(lastKind))),
		},
	}

	r.GORMModel = r.REST.SingularName + "M"
	r.MapModelToAPIFunc = fmt.Sprintf("%sMTo%s%s", r.REST.SingularName, r.REST.SingularName, upperVer)
	r.MapAPIToModelFunc = fmt.Sprintf("%s%sTo%sM", r.REST.SingularName, upperVer, r.REST.SingularName)
	r.BusinessFactoryName = fmt.Sprintf("%s%s", r.REST.SingularName, upperVer)
	r.ResourcePathPrefix = strings.ToLower(filepath.Dir(kindPath))
	if r.ResourcePathPrefix == "." {
		r.ResourcePathPrefix = ""
	}
	r.FileName = r.REST.SingularLower + ".go"
	if r.ResourcePathPrefix != "" {
		r.FileName = strings.ReplaceAll(r.ResourcePathPrefix, "/", "_") + "_" + r.FileName
	}

	return &r
}

// SetRESTGen attaches REST metadata for later template rendering.
func (ws *WebServer) SetRESTGen(meta *RESTGen) *WebServer {
	ws.R = meta
	return ws
}

func (ws *WebServer) GetR() *RESTGen {
	return ws.R
}

func (ws *WebServer) GetProj() *Project {
	return ws.Proj
}

func (ws *WebServer) GetClients() []string {
	return ws.Clients
}

// Pairs returns a map of destination relative paths to template paths.
// It drives file generation for this component.
func (ws *WebServer) Pairs() map[string]string {
	// Local shortcuts to reduce repetition.
	apiDir := filepath.Join("pkg/api", ws.Name, ws.Proj.M.APIVersion)
	internalPkg := ws.Proj.InternalPkg()
	baseDir := ws.BaseDir()
	handlerDir := ws.HandlerDir()
	storeDir := ws.StoreDir()
	bizDir := ws.BizDir()
	pkgDir := ws.PkgDir()

	pairs := map[string]string{}
	add := func(dst, tpl string) {
		pairs[dst] = tpl
	}

	// Common command and component scaffolding.
	add(filepath.Join("cmd", ws.BinaryName, "app/options/options.go"), "/project/cmd/mb-apiserver/app/options/options.go")
	add(filepath.Join("cmd", ws.BinaryName, "app/server.go"), "/project/cmd/mb-apiserver/app/server.go")
	add(filepath.Join("cmd", ws.BinaryName, "main.go"), "/project/cmd/mb-apiserver/main.go")
	add(filepath.Join(ws.Proj.Configs(), ws.BinaryName+".yaml"), "/project/configs/mb-apiserver.yaml")

	add(filepath.Join(ws.ModelDir(), "fake.go"), "/project/internal/apiserver/model/fake.go")

	// Core internal packages reused across frameworks.
	add(filepath.Join(storeDir, "doc.go"), "/project/internal/apiserver/store/doc.go")
	add(filepath.Join(storeDir, "store.go"), "/project/internal/apiserver/store/store.go")
	add(filepath.Join(storeDir, "fake.go"), "/project/internal/apiserver/store/fake.go")
	add(filepath.Join(storeDir, "README.md"), "/project/internal/apiserver/store/README.md")

	add(filepath.Join(bizDir, "biz.go"), "/project/internal/apiserver/biz/biz.go")
	add(filepath.Join(bizDir, "doc.go"), "/project/internal/apiserver/biz/doc.go")
	add(filepath.Join(bizDir, "README.md"), "/project/internal/apiserver/biz/README.md")

	add(filepath.Join(ws.PkgDir(), "validation/validation.go"), "/project/internal/apiserver/pkg/validation/validation.go")

	add(filepath.Join(baseDir, "server.go"), "/project/internal/apiserver/server.go")
	add(filepath.Join(baseDir, "wire.go"), "/project/internal/apiserver/wire.go")
	add(filepath.Join(baseDir, "wire_gen.go"), "/project/internal/apiserver/wire_gen.go")

	// Default proto for examples.
	add(filepath.Join(apiDir, "example.proto"), "/project/pkg/api/apiserver/v1/example.proto")

	AddGenericPackages(pairs, internalPkg)
	GenerateKubernetesManifests(pairs, ws.Proj.Metadata, ws.BinaryName)

	// Optional 'user' feature.
	if ws.WithUser {
		ws.PrepareRESTMetadata("user")

		add(filepath.Join(apiDir, "user.proto"), "/project/pkg/api/apiserver/v1/user.proto")
		add(filepath.Join(internalPkg, "known/role.go"), "/project/internal/pkg/known/role.go")
		add(filepath.Join(internalPkg, "errno/user.go"), "/project/internal/pkg/errno/user.go")

		// Model
		add(filepath.Join(ws.ModelDir(), "user.gen.go"), "/project/internal/apiserver/model/user.gen.go")
		add(filepath.Join(ws.ModelDir(), "hook_user.go"), "/project/internal/apiserver/model/hook_user.go")

		// Handler + middlewares by framework
		switch ws.WebFramework {
		case known.WebFrameworkGin:
			add(filepath.Join(handlerDir, "user.go"), "/project/internal/apiserver/handler/gin/user.go")
			add(filepath.Join(internalPkg, "middleware/gin/authn.go"), "/project/internal/pkg/middleware/gin/authn.go")
			add(filepath.Join(internalPkg, "middleware/gin/authz.go"), "/project/internal/pkg/middleware/gin/authz.go")
		case known.WebFrameworkGRPC:
			add(filepath.Join(handlerDir, "user.go"), "/project/internal/apiserver/handler/grpc/user.go")
			add(filepath.Join(internalPkg, "middleware/grpc/authn.go"), "/project/internal/pkg/middleware/grpc/authn.go")
			add(filepath.Join(internalPkg, "middleware/grpc/authz.go"), "/project/internal/pkg/middleware/grpc/authz.go")
			add(filepath.Join("examples/client/user/main.go"), "/project/examples/client/user/main.go")
			add(filepath.Join("examples/helper/helper.go"), "/project/examples/helper/helper.go")
			add(filepath.Join("examples/helper/README.md"), "/project/examples/helper/README.md")
		}

		// Conversion/validation
		add(filepath.Join(ws.PkgDir(), "conversion/user.go"), "/project/internal/apiserver/pkg/conversion/user.go")
		add(filepath.Join(ws.PkgDir(), "validation/user.go"), "/project/internal/apiserver/pkg/validation/user.go")

		// Biz + store
		add(ws.RESTBizFile(), "/project/internal/apiserver/biz/v1/user/user.go")
		add(ws.RESTStoreFile(), "/project/internal/apiserver/store/user.go")
	}

	// Optional healthz endpoints.
	if ws.WithHealthz {
		add(filepath.Join(apiDir, "healthz.proto"), "/project/pkg/api/apiserver/v1/healthz.proto")

		switch ws.WebFramework {
		case known.WebFrameworkGin:
			add(filepath.Join(handlerDir, "healthz.go"), "/project/internal/apiserver/handler/gin/healthz.go")
		case known.WebFrameworkGRPC:
			add(filepath.Join(handlerDir, "healthz.go"), "/project/internal/apiserver/handler/grpc/healthz.go")
			add(filepath.Join("examples/client/health/main.go"), "/project/examples/client/health/main.go")
		}
	}

	if ws.WithOTel {
		add(filepath.Join(pkgDir, "metrics/metrics.go"), "/project/internal/apiserver/pkg/metrics/metrics.go")
	}

	if ws.WithWS {
		add(filepath.Join(apiDir, "wsmessage.proto"), "/project/pkg/api/apiserver/v1/wsmessage.proto")
		add(filepath.Join(internalPkg, "errno/websocket.go"), "/project/internal/pkg/errno/websocket.go")
		add(filepath.Join(handlerDir, "websocket.go"), "/project/internal/apiserver/handler/gin/websocket.go")
		add(filepath.Join(
			bizDir,
			ws.Proj.M.APIVersion,
			"websocket/websocket.go"),
			"/project/internal/apiserver/biz/v1/websocket/websocket.go",
		)
		add(filepath.Join(ws.Proj.Examples(), "websocket/ws-client.go"), "/project/examples/websocket/ws-client.go")
		add(filepath.Join(ws.Proj.Examples(), "websocket/test-websocket.sh"), "/project/examples/websocket/test-websocket.sh")
	}

	if ws.WithPreloader {
		add(filepath.Join(pkgDir, "asyncstore/asyncstore.go"), "/project/internal/apiserver/pkg/asyncstore/asyncstore.go")
		add(filepath.Join(pkgDir, "asyncstore/fake_store.go"), "/project/internal/apiserver/pkg/asyncstore/fake_store.go")
		add(filepath.Join(apiDir, "fake.proto"), "/project/pkg/api/apiserver/v1/fake.proto")
	}

	// Framework-specific scaffolding.
	switch ws.WebFramework {
	case known.WebFrameworkGin:
		add(filepath.Join(internalPkg, "middleware/gin/header.go"), "/project/internal/pkg/middleware/gin/header.go")
		add(filepath.Join(internalPkg, "middleware/gin/requestid.go"), "/project/internal/pkg/middleware/gin/requestid.go")
		add(filepath.Join(baseDir, "httpserver.go"), "/project/internal/apiserver/ginserver.go")
		add(filepath.Join(handlerDir, "handler.go"), "/project/internal/apiserver/handler/gin/handler.go")
		add(filepath.Join(ws.Proj.Scripts(), "startup-test.sh"), "/project/scripts/startup-test.sh")
		if ws.WithOTel {
			add(filepath.Join(internalPkg, "middleware/gin/context.go"), "/project/internal/pkg/middleware/gin/context.go")
		}
	case known.WebFrameworkGRPC:
		// grpc middlewares
		add(filepath.Join(internalPkg, "middleware/grpc/requestid.go"), "/project/internal/pkg/middleware/grpc/requestid.go")
		add(filepath.Join(internalPkg, "middleware/grpc/doc.go"), "/project/internal/pkg/middleware/grpc/doc.go")
		add(filepath.Join(internalPkg, "middleware/grpc/defaulter.go"), "/project/internal/pkg/middleware/grpc/defaulter.go")
		add(filepath.Join(internalPkg, "middleware/grpc/validator.go"), "/project/internal/pkg/middleware/grpc/validator.go")

		// apiserver proto and servers
		add(filepath.Join(apiDir, ws.Name+".proto"), "/project/pkg/api/apiserver/v1/apiserver.proto")
		switch ws.ServiceRegistry {
		case known.ServiceRegistryPolaris:
			add(filepath.Join(baseDir, "grpcserver.go"), "/project/internal/apiserver/polarisserver.go")
		default:
			add(filepath.Join(baseDir, "grpcserver.go"), "/project/internal/apiserver/grpcserver.go")
		}
		add(filepath.Join(handlerDir, "handler.go"), "/project/internal/apiserver/handler/grpc/handler.go")
		if ws.WithOTel {
			add(filepath.Join(internalPkg, "middleware/grpc/context.go"), "/project/internal/pkg/middleware/grpc/context.go")
		}
	case known.WebFrameworkGRPCGateway:
		// TODO: add grpc-gateway templates if needed

	case known.WebFrameworkKratos:
		// TODO: add kratos templates if needed

	default:
		// Fallback to gRPC server scaffolding.
		add(filepath.Join(apiDir, ws.Name+".proto"), "/project/pkg/api/apiserver/v1/apiserver.proto")
		add(filepath.Join(baseDir, "grpcserver.go"), "/project/internal/apiserver/grpcserver.go")
		add(filepath.Join(handlerDir, "handler.go"), "/project/internal/apiserver/handler/grpc/handler.go")
	}

	if len(ws.Clients) > 0 {
		add(filepath.Join(ws.PkgDir(), "clientset/clientset.go"), "/project/internal/apiserver/pkg/clientset/clientset.go")
	}

	// Ensure api dir exists in VCS.
	add(filepath.Join("api/.keep"), "/keep.tpl")

	return pairs
}
