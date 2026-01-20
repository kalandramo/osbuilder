package types

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/duke-git/lancet/v2/strutil"
	"github.com/gobuffalo/flect"
	stringsutil "github.com/onexstack/onexstack/pkg/util/strings"

	"github.com/onexstack/osbuilder/internal/osbuilder/known"
)

// MQServer describes a mq server component to generate (HTTP/gRPC/etc).
type MQServer struct {
	// BinaryName is the CLI binary name (e.g., "mb-apiserver").
	BinaryName string `yaml:"binaryName"`
	// MQFramework selects the framework (e.g., gin, grpc).
	MQFramework string `yaml:"mqFramework"`
	// StorageType selects backing storage (e.g., memory, mysql).
	StorageType string `yaml:"storageType"`
	// Feature flags
	WithHealthz   bool     `yaml:"withHealthz,omitempty"`
	WithOTel      bool     `yaml:"withOTel,omitempty"`
	WithPreloader bool     `yaml:"withPreloader,omitempty"`
	Clients       []string `yaml:"clients,omitempty"`

	// Computed/derived fields (not serialized).
	ServerGen `yaml:"-"`
}

// Complete populates derived fields and sensible defaults.
func (mq *MQServer) Complete(proj *Project) *MQServer {
	mq.ServerGen = NewServerGen(proj, mq.BinaryName)
	return mq
}

// HandlerDir returns the handler directory for the component.
func (mq *MQServer) HandlerDir() string {
	return filepath.Join(mq.BaseDir(), "handler")
}

// BizDir returns the business logic directory for the component.
func (mq *MQServer) BizDir() string {
	return filepath.Join(mq.BaseDir(), "biz")
}

// RESTBizFile returns the business logic directory for the specified rest resource.
func (mq *MQServer) RESTBizFile() string {
	return filepath.Join(
		mq.BizDir(),
		mq.Proj.M.APIVersion,
		mq.R.ResourcePathPrefix,
		mq.R.REST.SingularLower,
		mq.R.FileName,
	)
}

// StoreDir returns the data store directory for the component.
func (mq *MQServer) StoreDir() string {
	return filepath.Join(mq.BaseDir(), "store")
}

// RESTStoreFile returns the path to a REST store implementation for a singular resource.
func (mq *MQServer) RESTStoreFile() string {
	return filepath.Join(mq.StoreDir(), mq.R.FileName)
}

// BaseDir returns the component base directory: internal/<component>.
func (mq *MQServer) BaseDir() string {
	return filepath.Join("internal", mq.Name)
}

// Model returns the model directory for the component.
func (mq *MQServer) ModelDir() string {
	return filepath.Join(mq.BaseDir(), "model")
}

// Pkg returns the pkg directory for the component.
func (mq *MQServer) PkgDir() string {
	return filepath.Join(mq.BaseDir(), "pkg")
}

// API returns the API directory for the component: pkg/api/<component>/<version>.
func (mq *MQServer) APIDir() string {
	return filepath.Join("pkg/api", mq.Name, mq.Proj.M.APIVersion)
}

// PrepareRESTMetadata constructs REST metadata for a given kind.
func (mq *MQServer) PrepareRESTMetadata(kindPath string) {
	lastKind := filepath.Base(kindPath)
	// kind := strings.ReplaceAll(kindPath, "/", "_")
	upperVer := strings.ToUpper(mq.Proj.M.APIVersion)

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

	mq.R = &r
}

// SetRESTGen attaches REST metadata for later template rendering.
func (mq *MQServer) SetRESTGen(meta *RESTGen) *MQServer {
	mq.R = meta
	return mq
}

// Pairs returns a map of destination relative paths to template paths.
// It drives file generation for this component.
func (mq *MQServer) Pairs() map[string]string {
	// Local shortcuts to reduce repetition.
	apiDir := filepath.Join("pkg/api", mq.Name, mq.Proj.M.APIVersion)
	internalPkg := mq.Proj.InternalPkg()
	baseDir := mq.BaseDir()
	handlerDir := mq.HandlerDir()
	storeDir := mq.StoreDir()
	bizDir := mq.BizDir()
	pkgDir := mq.PkgDir()

	pairs := map[string]string{}
	add := func(dst, tpl string) {
		pairs[dst] = tpl
	}

	// Common command and component scaffolding.
	add(filepath.Join("cmd", mq.BinaryName, "app/options/options.go"), "/project/cmd/mb-apiserver/app/options/options.go")
	add(filepath.Join("cmd", mq.BinaryName, "app/server.go"), "/project/cmd/mb-apiserver/app/server.go")
	add(filepath.Join("cmd", mq.BinaryName, "main.go"), "/project/cmd/mb-apiserver/main.go")
	add(filepath.Join(mq.Proj.Configs(), mq.BinaryName+".yaml"), "/project/configs/mb-apiserver.yaml")

	// Core internal packages reused across frameworks.
	add(filepath.Join(storeDir, "doc.go"), "/project/internal/apiserver/store/doc.go")
	add(filepath.Join(storeDir, "store.go"), "/project/internal/apiserver/store/store.go")
	add(filepath.Join(storeDir, "README.md"), "/project/internal/apiserver/store/README.md")

	add(filepath.Join(bizDir, "biz.go"), "/project/internal/apiserver/biz/biz.go")
	add(filepath.Join(bizDir, "doc.go"), "/project/internal/apiserver/biz/doc.go")
	add(filepath.Join(bizDir, "README.md"), "/project/internal/apiserver/biz/README.md")

	add(filepath.Join(mq.PkgDir(), "validation/validation.go"), "/project/internal/apiserver/pkg/validation/validation.go")

	add(filepath.Join(internalPkg, "contextx/contextx.go"), "/project/internal/pkg/contextx/contextx.go")
	add(filepath.Join(internalPkg, "contextx/doc.go"), "/project/internal/pkg/contextx/doc.go")
	add(filepath.Join(internalPkg, "known/doc.go"), "/project/internal/pkg/known/doc.go")
	add(filepath.Join(internalPkg, "known/known.go"), "/project/internal/pkg/known/known.go")

	add(filepath.Join(internalPkg, "rid/doc.go"), "/project/internal/pkg/rid/doc.go")
	add(filepath.Join(internalPkg, "rid/example_test.go"), "/project/internal/pkg/rid/example_test.go")
	add(filepath.Join(internalPkg, "rid/rid.go"), "/project/internal/pkg/rid/rid.go")
	add(filepath.Join(internalPkg, "rid/rid_test.go"), "/project/internal/pkg/rid/rid_test.go")
	add(filepath.Join(internalPkg, "rid/salt.go"), "/project/internal/pkg/rid/salt.go")

	add(filepath.Join(internalPkg, "errno/doc.go"), "/project/internal/pkg/errno/doc.go")
	add(filepath.Join(internalPkg, "errno/code.go"), "/project/internal/pkg/errno/code.go")

	add(filepath.Join(baseDir, "server.go"), "/project/internal/apiserver/server.go")
	add(filepath.Join(baseDir, "wire.go"), "/project/internal/apiserver/wire.go")
	add(filepath.Join(baseDir, "wire_gen.go"), "/project/internal/apiserver/wire_gen.go")

	// Default proto for examples.
	add(filepath.Join(apiDir, "example.proto"), "/project/pkg/api/apiserver/v1/example.proto")

	if stringsutil.StringIn(mq.Proj.Metadata.DeploymentMethod, []string{known.DeploymentModeDocker, known.DeploymentModeKubernetes}) {
		switch mq.Proj.Metadata.Image.DockerfileMode {
		case known.DockerfileModeNone:
			add(filepath.Join("build", "docker", mq.BinaryName, ".keep"), "/keep.tpl")
		case known.DockerfileModeRuntimeOnly:
			add(filepath.Join("build", "docker", mq.BinaryName, "Dockerfile"), "/project/build/docker/mb-apiserver/Dockerfile.runtime-only")
		case known.DockerfileModeMultiStage:
			add(filepath.Join("build", "docker", mq.BinaryName, "Dockerfile"), "/project/build/docker/mb-apiserver/Dockerfile.multi-stage")
		case known.DockerfileModeCombined:
			add(filepath.Join("build", "docker", mq.BinaryName, "Dockerfile"), "/project/build/docker/mb-apiserver/Dockerfile.multi-stage")
			add(filepath.Join("build", "docker", mq.BinaryName, "Dockerfile.runtime-only"), "/project/build/docker/mb-apiserver/Dockerfile.runtime-only")
		default:
		}
	}

	if mq.Proj.Metadata.DeploymentMethod == known.DeploymentModeKubernetes {
		add(filepath.Join("manifests", mq.BinaryName, mq.BinaryName+".deployment.yaml"), "/project/manifests/mb-apiserver/mb-apiserver.deployment.yaml")
		add(filepath.Join("manifests", mq.BinaryName, mq.BinaryName+".service.yaml"), "/project/manifests/mb-apiserver/mb-apiserver.service.yaml")
		add(filepath.Join("manifests", mq.BinaryName, mq.BinaryName+".configmap.yaml"), "/project/manifests/mb-apiserver/mb-apiserver.configmap.yaml")
		add(filepath.Join("manifests", "nettool.deployment.yaml"), "/project/manifests/nettool.deployment.yaml")
		add(filepath.Join("manifests", "nginx.deployment.yaml"), "/project/manifests/nginx.deployment.yaml")
	}

	// Optional healthz endpoints.
	if mq.WithHealthz {
		add(filepath.Join(apiDir, "healthz.proto"), "/project/pkg/api/apiserver/v1/healthz.proto")
	}

	if mq.WithOTel {
		add(filepath.Join(pkgDir, "metrics/metrics.go"), "/project/internal/apiserver/pkg/metrics/metrics.go")
	}

	if mq.WithPreloader {
		add(filepath.Join(pkgDir, "asyncstore/asyncstore.go"), "/project/internal/apiserver/pkg/asyncstore/asyncstore.go")
		add(filepath.Join(pkgDir, "asyncstore/fake_store.go"), "/project/internal/apiserver/pkg/asyncstore/fake_store.go")
		add(filepath.Join(apiDir, "fake.proto"), "/project/pkg/api/apiserver/v1/fake.proto")
	}

	// Framework-specific scaffolding.
	switch mq.MQFramework {
	case known.MQFrameworkKafka:
		add(filepath.Join(internalPkg, "middleware/gin/header.go"), "/project/internal/pkg/middleware/gin/header.go")
		add(filepath.Join(internalPkg, "middleware/gin/requestid.go"), "/project/internal/pkg/middleware/gin/requestid.go")
		add(filepath.Join(baseDir, "httpserver.go"), "/project/internal/apiserver/ginserver.go")
		add(filepath.Join(handlerDir, "handler.go"), "/project/internal/apiserver/handler/gin/handler.go")
		add(filepath.Join(mq.Proj.Scripts(), "startup-test.sh"), "/project/scripts/startup-test.sh")
		if mq.WithOTel {
			add(filepath.Join(internalPkg, "middleware/gin/context.go"), "/project/internal/pkg/middleware/gin/context.go")
		}
	default:
		// Fallback to gRPC server scaffolding.
		add(filepath.Join(apiDir, mq.Name+".proto"), "/project/pkg/api/apiserver/v1/apiserver.proto")
		add(filepath.Join(baseDir, "grpcserver.go"), "/project/internal/apiserver/grpcserver.go")
		add(filepath.Join(handlerDir, "handler.go"), "/project/internal/apiserver/handler/grpc/handler.go")
	}

	if len(mq.Clients) > 0 {
		add(filepath.Join(mq.PkgDir(), "clientset/clientset.go"), "/project/internal/apiserver/pkg/clientset/clientset.go")
	}

	// Ensure api dir exists in VCS.
	add(filepath.Join("api/.keep"), "/keep.tpl")

	return pairs
}
