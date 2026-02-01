package types

import (
	"path/filepath"
)

type WebServerS struct {
	BinaryName string `yaml:"binaryName"`
	Name       string `yaml:"-"`
}

// JobServer describes a job server component to generate (HTTP/gRPC/etc).
type JobServer struct {
	// BinaryName is the CLI binary name (e.g., "mb-jobserver").
	BinaryName string `yaml:"binaryName"`
	// StorageType selects backing storage (e.g., memory, mysql).
	StorageType string `yaml:"storageType"`
	// Feature flags
	// WithHealthz   bool     `yaml:"withHealthz,omitempty"`
	DisableCronJob bool     `yaml:"disableCronJob,omitempty"`
	WithOTel       bool     `yaml:"withOTel,omitempty"`
	WithPreloader  bool     `yaml:"withPreloader,omitempty"`
	Clients        []string `yaml:"clients,omitempty"`

	// Computed/derived fields (not serialized).
	ServerGen `yaml:"-"`
}

// Complete populates derived fields and sensible defaults.
func (job *JobServer) Complete(proj *Project) *JobServer {
	job.ServerGen = NewServerGen(proj, job.BinaryName)
	return job
}

// HandlerDir returns the handler directory for the component.
func (job *JobServer) HandlerDir() string {
	return filepath.Join(job.BaseDir(), "handler")
}

// BizDir returns the business logic directory for the component.
func (job *JobServer) BizDir() string {
	return filepath.Join(job.BaseDir(), "biz")
}

// RESTBizFile returns the business logic directory for the specified rest resource.
func (job *JobServer) RESTBizFile() string {
	return filepath.Join(
		job.BizDir(),
		job.Proj.M.APIVersion,
		job.R.ResourcePathPrefix,
		job.R.REST.SingularLower,
		job.R.REST.SingularLower+".go",
	)
}

// StoreDir returns the data store directory for the component.
func (job *JobServer) StoreDir() string {
	return filepath.Join(job.BaseDir(), "store")
}

// RESTStoreFile returns the path to a REST store implementation for a singular resource.
func (job *JobServer) RESTStoreFile() string {
	return filepath.Join(job.StoreDir(), job.R.FileName)
}

// BaseDir returns the component base directory: internal/<component>.
func (job *JobServer) BaseDir() string {
	return filepath.Join("internal", job.Name)
}

func (job *JobServer) WatcherDir() string {
	return filepath.Join(job.BaseDir(), "watcher")
}

// Model returns the model directory for the component.
func (job *JobServer) ModelDir() string {
	return filepath.Join(job.BaseDir(), "model")
}

// Pkg returns the pkg directory for the component.
func (job *JobServer) PkgDir() string {
	return filepath.Join(job.BaseDir(), "pkg")
}

// API returns the API directory for the component: pkg/api/<component>/<version>.
func (job *JobServer) APIDir() string {
	return filepath.Join("pkg/api", job.Name, job.Proj.M.APIVersion)
}

// PrepareRESTMetadata constructs REST metadata for a given kind.
func (job *JobServer) PrepareRESTMetadata(kindPath string) {
	job.R = prepareRESTMetadata(job.Proj.M.APIVersion, kindPath)
}

// SetRESTGen attaches REST metadata for later template rendering.
func (job *JobServer) SetRESTGen(meta *RESTGen) *JobServer {
	job.R = meta
	return job
}

func (job *JobServer) GetR() *RESTGen {
	return job.R
}

func (job *JobServer) GetProj() *Project {
	return job.Proj
}

func (job *JobServer) GetClients() []string {
	return job.Clients
}

// Pairs returns a map of destination relative paths to template paths.
// It drives file generation for this component.
func (job *JobServer) Pairs() map[string]string {
	// Local shortcuts to reduce repetition.
	apiDir := filepath.Join("pkg/api", job.Name, job.Proj.M.APIVersion)
	internalPkg := job.Proj.InternalPkg()
	baseDir := job.BaseDir()
	watcherDir := job.WatcherDir()
	modelDir := job.ModelDir()
	// handlerDir := job.HandlerDir()
	storeDir := job.StoreDir()
	// bizDir := job.BizDir()
	pkgDir := job.PkgDir()

	pairs := map[string]string{}
	add := func(dst, tpl string) {
		pairs[dst] = tpl
	}

	// Common command and component scaffolding.
	add(filepath.Join("cmd", job.BinaryName, "main.go"), "/project/cmd/mb-jobserver/main.go")
	add(filepath.Join("cmd", job.BinaryName, "app/server.go"), "/project/cmd/mb-jobserver/app/server.go")
	add(filepath.Join("cmd", job.BinaryName, "app/options/options.go"), "/project/cmd/mb-jobserver/app/options/options.go")

	add(filepath.Join(baseDir, "wire.go"), "/project/internal/jobserver/wire.go")
	add(filepath.Join(baseDir, "server.go"), "/project/internal/jobserver/server.go")
	add(filepath.Join(baseDir, "jobserver.go"), "/project/internal/jobserver/jobserver.go")
	add(filepath.Join(baseDir, "wire_gen.go"), "/project/internal/jobserver/wire_gen.go")

	add(filepath.Join(modelDir, "fake.go"), "/project/internal/jobserver/model/fake.go")

	add(filepath.Join(storeDir, "doc.go"), "/project/internal/jobserver/store/doc.go")
	add(filepath.Join(storeDir, "store.go"), "/project/internal/jobserver/store/store.go")
	add(filepath.Join(storeDir, "fake.go"), "/project/internal/jobserver/store/fake.go")
	add(filepath.Join(storeDir, "README.md"), "/project/internal/jobserver/store/README.md")

	if job.WithOTel {
		add(filepath.Join(pkgDir, "metrics/metrics.go"), "/project/internal/jobserver/pkg/metrics/metrics.go")
	}
	if job.WithPreloader {
		add(filepath.Join(pkgDir, "asyncstore/asyncstore.go"), "/project/internal/jobserver/pkg/asyncstore/asyncstore.go")
		add(filepath.Join(pkgDir, "asyncstore/fake_store.go"), "/project/internal/jobserver/pkg/asyncstore/fake_store.go")
		add(filepath.Join(apiDir, "fake.proto"), "/project/pkg/api/jobserver/v1/fake.proto")
	}
	if len(job.Clients) > 0 {
		add(filepath.Join(pkgDir, "clientset/clientset.go"), "/project/internal/jobserver/pkg/clientset/clientset.go")
	}

	// Add watchers
	add(filepath.Join(watcherDir, "config.go"), "/project/internal/jobserver/watcher/config.go")
	add(filepath.Join(watcherDir, "initializer.go"), "/project/internal/jobserver/watcher/initializer.go")
	add(filepath.Join(watcherDir, "interface.go"), "/project/internal/jobserver/watcher/interface.go")
	add(filepath.Join(watcherDir, "all/all.go"), "/project/internal/jobserver/watcher/all/all.go")

	if !job.DisableCronJob {
		add(filepath.Join(apiDir, "cronjob.proto"), "/project/pkg/api/jobserver/v1/cronjob.proto")
		add(filepath.Join(apiDir, "job.proto"), "/project/pkg/api/jobserver/v1/job.proto")

		add(filepath.Join(modelDir, "cronjob.go"), "/project/internal/jobserver/model/cronjob.go")
		add(filepath.Join(modelDir, "hook_cronjob.go"), "/project/internal/jobserver/model/hook_cronjob.go")
		add(filepath.Join(modelDir, "job.go"), "/project/internal/jobserver/model/job.go")
		add(filepath.Join(modelDir, "hook_job.go"), "/project/internal/jobserver/model/hook_job.go")
		add(filepath.Join(modelDir, "model.go"), "/project/internal/jobserver/model/model.go")
		add(filepath.Join(storeDir, "cronjob.go"), "/project/internal/jobserver/store/cronjob.go")
		add(filepath.Join(storeDir, "job.go"), "/project/internal/jobserver/store/job.go")

		add(filepath.Join(pkgDir, "path/path.go"), "/project/internal/jobserver/pkg/path/path.go")
		add(filepath.Join(pkgDir, "clientset/typed/fakeminio/fakeminio.go"), "/project/internal/jobserver/pkg/clientset/typed/fakeminio/fakeminio.go")
		add(filepath.Join(pkgDir, "clientset/typed/minio/minio.go"), "/project/internal/jobserver/pkg/clientset/typed/minio/minio.go")
		add(filepath.Join(pkgDir, "clientset/typed/train/train.go"), "/project/internal/jobserver/pkg/clientset/typed/train/train.go")
		add(filepath.Join(watcherDir, "cronjob/cronjob/history.go"), "/project/internal/jobserver/watcher/cronjob/cronjob/history.go")
		add(filepath.Join(watcherDir, "cronjob/cronjob/watcher.go"), "/project/internal/jobserver/watcher/cronjob/cronjob/watcher.go")
		add(filepath.Join(watcherDir, "cronjob/statesync/watcher.go"), "/project/internal/jobserver/watcher/cronjob/statesync/watcher.go")
		add(filepath.Join(watcherDir, "job/llmtrain/event.go"), "/project/internal/jobserver/watcher/job/llmtrain/event.go")
		add(filepath.Join(watcherDir, "job/llmtrain/fsm.go"), "/project/internal/jobserver/watcher/job/llmtrain/fsm.go")
		add(filepath.Join(watcherDir, "job/llmtrain/helper.go"), "/project/internal/jobserver/watcher/job/llmtrain/helper.go")
		add(filepath.Join(watcherDir, "job/llmtrain/watcher.go"), "/project/internal/jobserver/watcher/job/llmtrain/watcher.go")
	}
	add(filepath.Join(watcherDir, "customized/fake/watcher.go"), "/project/internal/jobserver/watcher/customized/fake/watcher.go")

	// Add other files
	if !job.DisableCronJob {
		add(filepath.Join(job.Proj.Configs(), "cronjob.sql"), "/project/configs/cronjob.sql")
	}
	add(filepath.Join(job.Proj.Configs(), job.BinaryName+".yaml"), "/project/configs/mb-jobserver.yaml")
	add(filepath.Join(job.Proj.Examples(), "jobserver/cronjob/insert_job.sh"), "/project/examples/jobserver/cronjob/insert_job.sh")

	add(filepath.Join(internalPkg, "embedding/embedder/fake/fake.go"), "/project/internal/pkg/embedding/embedder/fake/fake.go")
	add(filepath.Join(internalPkg, "known/job/doc.go"), "/project/internal/pkg/known/job/doc.go")
	add(filepath.Join(internalPkg, "known/job/job.go"), "/project/internal/pkg/known/job/job.go")
	add(filepath.Join(internalPkg, "known/job/llmtrain.go"), "/project/internal/pkg/known/job/llmtrain.go")
	add(filepath.Join(internalPkg, "util/fsm/fsm.go"), "/project/internal/pkg/util/fsm/fsm.go")
	add(filepath.Join(internalPkg, "util/jobconditions/getter.go"), "/project/internal/pkg/util/jobconditions/getter.go")
	add(filepath.Join(internalPkg, "util/jobconditions/setter.go"), "/project/internal/pkg/util/jobconditions/setter.go")

	// Core internal packages reused across frameworks.
	AddGenericPackages(pairs, internalPkg)
	GenerateKubernetesManifests(pairs, job.Proj.Metadata, job.BinaryName)

	return pairs
}
