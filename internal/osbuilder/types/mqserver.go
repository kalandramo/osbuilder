package types

import (
	"path/filepath"

	"github.com/onexstack/osbuilder/internal/osbuilder/known"
)

// MQServer describes a mq server component to generate (HTTP/gRPC/etc).
type MQServer struct {
	// BinaryName is the CLI binary name (e.g., "mb-mqserver").
	BinaryName string `yaml:"binaryName"`
	// MQFramework selects the framework (e.g., gin, grpc).
	MQFramework string `yaml:"mqFramework"`
	// StorageType selects backing storage (e.g., memory, mysql).
	StorageType string `yaml:"storageType"`
	// Feature flags
	// WithHealthz   bool     `yaml:"withHealthz,omitempty"`
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
		mq.R.REST.SingularLower+".go",
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
	mq.R = prepareRESTMetadata(mq.Proj.M.APIVersion, kindPath)
}

// SetRESTGen attaches REST metadata for later template rendering.
func (mq *MQServer) SetRESTGen(meta *RESTGen) *MQServer {
	mq.R = meta
	return mq
}

func (mq *MQServer) GetR() *RESTGen {
	return mq.R
}

func (mq *MQServer) GetProj() *Project {
	return mq.Proj
}

func (mq *MQServer) GetClients() []string {
	return mq.Clients
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
	modelDir := mq.ModelDir()
	bizDir := mq.BizDir()
	pkgDir := mq.PkgDir()

	pairs := map[string]string{}
	add := func(dst, tpl string) {
		pairs[dst] = tpl
	}

	// Common command and component scaffolding.
	add(filepath.Join("cmd", mq.BinaryName, "main.go"), "/project/cmd/mb-mqserver/main.go")
	add(filepath.Join("cmd", mq.BinaryName, "app/server.go"), "/project/cmd/mb-mqserver/app/server.go")
	add(filepath.Join("cmd", mq.BinaryName, "app/options/options.go"), "/project/cmd/mb-mqserver/app/options/options.go")

	add(filepath.Join(mq.Proj.Configs(), mq.BinaryName+".yaml"), "/project/configs/mb-mqserver.yaml")

	add(filepath.Join(modelDir, "fake.go"), "/project/internal/mqserver/model/fake.go")

	// Core internal packages reused across frameworks.
	add(filepath.Join(storeDir, "doc.go"), "/project/internal/mqserver/store/doc.go")
	add(filepath.Join(storeDir, "store.go"), "/project/internal/mqserver/store/store.go")
	add(filepath.Join(storeDir, "fake.go"), "/project/internal/mqserver/store/fake.go")
	add(filepath.Join(storeDir, "README.md"), "/project/internal/mqserver/store/README.md")

	add(filepath.Join(bizDir, "biz.go"), "/project/internal/mqserver/biz/biz.go")
	add(filepath.Join(bizDir, "doc.go"), "/project/internal/mqserver/biz/doc.go")
	add(filepath.Join(bizDir, "README.md"), "/project/internal/mqserver/biz/README.md")

	add(filepath.Join(mq.PkgDir(), "validation/validation.go"), "/project/internal/mqserver/pkg/validation/validation.go")

	add(filepath.Join(internalPkg, "kafka/kafka.go"), "/project/internal/pkg/kafka/kafka.go")

	add(filepath.Join(baseDir, "server.go"), "/project/internal/mqserver/server.go")
	add(filepath.Join(baseDir, "wire.go"), "/project/internal/mqserver/wire.go")
	add(filepath.Join(baseDir, "wire_gen.go"), "/project/internal/mqserver/wire_gen.go")

	// Default proto for examples.
	add(filepath.Join(apiDir, "example.proto"), "/project/pkg/api/mqserver/v1/example.proto")

	AddGenericPackages(pairs, internalPkg)
	GenerateKubernetesManifests(pairs, mq.Proj.Metadata, mq.BinaryName)

	/*
		// Optional healthz endpoints.
		if mq.WithHealthz {
			add(filepath.Join(apiDir, "healthz.proto"), "/project/pkg/api/mqserver/v1/healthz.proto")
		}
	*/

	if mq.WithOTel {
		add(filepath.Join(pkgDir, "metrics/metrics.go"), "/project/internal/mqserver/pkg/metrics/metrics.go")
	}

	if mq.WithPreloader {
		add(filepath.Join(pkgDir, "asyncstore/asyncstore.go"), "/project/internal/mqserver/pkg/asyncstore/asyncstore.go")
		add(filepath.Join(pkgDir, "asyncstore/fake_store.go"), "/project/internal/mqserver/pkg/asyncstore/fake_store.go")
		add(filepath.Join(apiDir, "fake.proto"), "/project/pkg/api/mqserver/v1/fake.proto")
	}

	// Framework-specific scaffolding.
	switch mq.MQFramework {
	case known.MQFrameworkKafka:
		add(filepath.Join(baseDir, "mqserver.go"), "/project/internal/mqserver/kafkaserver.go")
		add(filepath.Join(handlerDir, "handler.go"), "/project/internal/mqserver/handler/kafka/handler.go")
		if mq.WithOTel {
			// add(filepath.Join(internalPkg, "middleware/gin/context.go"), "/project/internal/pkg/middleware/gin/context.go")
		}
	default:
		add(filepath.Join(baseDir, "mqserver.go"), "/project/internal/mqserver/kafkaserver.go.go")
		add(filepath.Join(handlerDir, "handler.go"), "/project/internal/mqserver/handler/kafka/handler.go")
	}

	if len(mq.Clients) > 0 {
		add(filepath.Join(mq.PkgDir(), "clientset/clientset.go"), "/project/internal/mqserver/pkg/clientset/clientset.go")
	}

	// Ensure api dir exists in VCS.
	add(filepath.Join("api/.keep"), "/keep.tpl")

	return pairs
}
