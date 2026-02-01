package types

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ImageConfig holds container image build configuration and Dockerfile generation options.
type ImageConfig struct {
	// RegistryPrefix is the image registry prefix where images are published.
	// (e.g., "docker.io/nginx", "gcr.io/distroless").
	RegistryPrefix string `yaml:"registryPrefix"`

	// DockerfileMode selects how Dockerfiles are generated for the project.
	// Supported values:
	// - "none": Do not generate a Dockerfile. You must provide your own at
	//   build/docker/<component_name>/Dockerfile.
	// - "runtime-only": Generate a runtime-only Dockerfile (expects an external build artifact).
	//   Useful for local debugging or when CI produces binaries/assets separately.
	// - "multi-stage": Generate a multi-stage Dockerfile (builder + runtime).
	// - "combined": Generate both Dockerfile variants:
	//   - Multi-stage: saved as "Dockerfile"
	//   - Runtime-only: saved as "Dockerfile.runtime-only"
	DockerfileMode string `yaml:"dockerfileMode"`

	// DistrolessMode define distroless usage modes.
	// - always:
	//   - forces distroless images for all environments (Docker builds).
	// - never:
	//   - forces standard distribution (dist) images (e.g., debian:bookworm) for all environments.
	// - auto:
	//   - Use distroless for production builds (e.g., multi-stage final stage).
	//   - Use dist images for non-production builds (e.g., runtime-only or dev targets) for easier debugging.
	DistrolessMode string `yaml:"distrolessMode"`
}

// Project is the top-level configuration for a generated project.
type Project struct {
	// Scaffold identifies the scaffold preset used.
	Scaffold string `yaml:"scaffold"`
	// Version tracks the schema or scaffold version.
	Version string `yaml:"version"`
	// Metadata provides project-level details (author, deployment, registry, etc.).
	Metadata *Metadata `yaml:"metadata"`

	// Use lowerCamelCase in YAML and pointer slices + omitempty for sparsity.
	WebServers []*WebServer `yaml:"webServers,omitempty"`
	JobServers []*JobServer `yaml:"jobServers,omitempty"`
	MQServers  []*MQServer  `yaml:"mqServers,omitempty"`
	CLITools   []*CLITool   `yaml:"cliTools,omitempty"`
	// DeclServers     []*DeclServer     `yaml:"declServers,omitempty"`
	// DeclControllers []*DeclController `yaml:"declControllers,omitempty"`

	// Common derived data used during generation (not serialized).
	M *ProjectGen `yaml:"-"`
}

// Metadata holds general project information and build/deploy preferences.
type Metadata struct {
	// ModulePath is the Go module import path as declared in the project's go.mod file.
	ModulePath string `yaml:"modulePath"`
	// Short is the short description shown in the 'help' output.
	ShortDescription string `yaml:"shortDescription"`
	// Long is the long message shown in the 'help <this-command>' output.
	LongMessage string `yaml:"longMessage"`
	// Image holds image build configuration.
	Image ImageConfig `yaml:"image"`
	// DeploymentMethod selects how to deploy (e.g., "kubernetes", "systemd").
	// Note: name kept for backward compatibility; often referred to as "deploymentMode".
	DeploymentMethod string `yaml:"deploymentMethod"`
	// MakefileMode selects the Makefile generation mode used by the project
	// (e.g., "none", "structed", "unstructured").
	MakefileMode string `yaml:"makefileMode"`
	// Author is the maintainer name.
	Author string `yaml:"author"`
	// Email is the maintainer email.
	Email string `yaml:"email"`
}

// FindWebServer returns the first web server whose BinaryName matches.
// If binaryName is empty, it returns the first available WebServer.
func (p *Project) FindWebServer(binaryName string) *WebServer {
	ws, _ := p.WebServerByBinary(binaryName)
	return ws
}

// WebServerByBinary returns a web server and a boolean indicating if it was found.
// If binaryName is empty, it defaults to the first non-nil WebServer in the list.
func (p *Project) WebServerByBinary(binaryName string) (*WebServer, bool) {
	for _, ws := range p.WebServers {
		// Skip nil entries to ensure safety
		if ws == nil {
			continue
		}

		// Logic:
		// 1. If binaryName is empty (""), return the first valid server encountered.
		// 2. Otherwise, return the server that strictly matches the binaryName.
		if binaryName == "" || ws.BinaryName == binaryName {
			return ws, true
		}
	}
	return nil, false
}

// FindMQServer returns the first web server whose BinaryName matches.
func (p *Project) FindMQServer(binaryName string) *MQServer {
	ws, _ := p.MQServerByBinary(binaryName)
	return ws
}

// MQServerByBinary returns a web server and a boolean indicating if it was found.
func (p *Project) MQServerByBinary(binaryName string) (*MQServer, bool) {
	for _, mq := range p.MQServers {
		// Skip nil entries to ensure safety
		if mq == nil {
			continue
		}

		if binaryName == "" || mq.BinaryName == binaryName {
			return mq, true
		}
	}
	return nil, false
}

// FindJobServer returns the first job server whose BinaryName matches.
func (p *Project) FindJobServer(binaryName string) *JobServer {
	job, _ := p.JobServerByBinary(binaryName)
	return job
}

// JobServerByBinary returns a web server and a boolean indicating if it was found.
func (p *Project) JobServerByBinary(binaryName string) (*JobServer, bool) {
	for _, job := range p.JobServers {
		// Skip nil entries to ensure safety
		if job == nil {
			continue
		}

		if binaryName == "" || job.BinaryName == binaryName {
			return job, true
		}
	}
	return nil, false
}

// FindCLITool returns the first job server whose BinaryName matches.
// Deprecated: use CLIToolByBinary to get an ok flag and avoid nil-struct ambiguity.
func (p *Project) FindCLITool(binaryName string) *CLITool {
	cli, _ := p.CLIToolByBinary(binaryName)
	return cli
}

// CLIToolByBinary returns a web server and a boolean indicating if it was found.
func (p *Project) CLIToolByBinary(binaryName string) (*CLITool, bool) {
	for _, cli := range p.CLITools {
		// Skip nil entries to ensure safety
		if cli == nil {
			continue
		}

		if binaryName == "" || cli.BinaryName == binaryName {
			return cli, true
		}
	}
	return nil, false
}

// Join builds a path under the project root (WorkDir).
// If WorkDir is empty, it joins the elements as-is.
func (p *Project) Join(elements ...string) string {
	base := ""
	if p != nil && p.M != nil {
		base = p.M.WorkDir
	}
	parts := elements
	if base != "" {
		parts = append([]string{base}, elements...)
	}
	return filepath.Join(parts...)
}

// Root returns the project root directory (WorkDir). Empty if not set.
func (p *Project) Root() string {
	if p == nil || p.M == nil {
		return ""
	}
	return p.M.WorkDir
}

// InternalPkg returns the path to the shared internal package directory.
func (p *Project) InternalPkg() string {
	return filepath.Join("internal", "pkg")
}

// Configs returns the path of configs.
func (p *Project) Configs() string {
	return "configs"
}

// Scripts returns the path to the scripts directory.
func (p *Project) Scripts() string {
	return filepath.Join("scripts")
}

// Examples returns the path to the examples directory.
func (p *Project) Examples() string {
	return filepath.Join("examples")
}

// Save writes the project to a YAML file at filename.
// It creates parent directories as needed and uses a 2-space indentation.
func (p *Project) Save(filename string) error {
	if filename == "" {
		return fmt.Errorf("filename must not be empty")
	}
	if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
		return fmt.Errorf("create parent dir: %w", err)
	}

	// Simple, reliable write with proper close semantics.
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("create file %q: %w", filename, err)
	}
	defer func() { _ = f.Close() }()

	// Header to prepend. It will be turned into YAML comments (`# ...`).
	header := `Copyright 2024 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
Use of this source code is governed by a MIT style
license that can be found in the LICENSE file. The original repo for
this file is https://github.com/onexstack/onex.

Code generated by osbuilder.
Only allow adding new components.
Modifying existing components or their configurations IS NOT ALLOWED.
Delete PROJECT or rename PROJECT IS NOT ALLOWED.`

	// Write header as YAML comments to keep the file a valid YAML document.
	// We prefix each line with "# ".
	commented := "# " + strings.ReplaceAll(header, "\n", "\n# ")
	if _, err := fmt.Fprintln(f, commented); err != nil {
		return fmt.Errorf("write header: %w", err)
	}

	enc := yaml.NewEncoder(f)
	enc.SetIndent(2)
	if err := enc.Encode(p); err != nil {
		_ = enc.Close()
		return fmt.Errorf("encode yaml: %w", err)
	}
	if err := enc.Close(); err != nil {
		return fmt.Errorf("close encoder: %w", err)
	}
	return nil
}

// REST captures the naming variants for the last/leaf segment of a resource path.
// For example, for a path like "ticket/cron_job", "cron_job" would be treated as the "last" segment.
type REST struct {
	// Singular form of the kind (e.g., "CronJob").
	SingularName string
	// Plural form of the kind (e.g., "CronJobs").
	PluralName string
	// Singular name in lower format (e.g., "cronjob").
	SingularLower string
	// Plural name in lower format (e.g., "cronjobs").
	PluralLower string
	// Singular name in lowerCamel (first letter lower) (e.g., "cronJob").
	SingularLowerFirst string
	// Plural name in lowerCamel (first letter lower) (e.g., "cronJobs").
	PluralLowerFirst string
}

// RESTGen captures naming conventions and helper metadata for generating REST resources.
type RESTGen struct {
	// REST holds the naming variants for the last (leaf) segment of the resource path,
	// such as the final component in "ticket/cron_job".
	REST
	// REST REST

	// Name of the associated GORM model (e.g., "CronJobModel").
	GORMModel string
	// Function name to map the model to the API.
	MapModelToAPIFunc string
	// Function name to map the API to the model.
	MapAPIToModelFunc string
	// Name of the business layer factory.
	BusinessFactoryName string

	// Name of the generated Go file.
	FileName string
	// ResourcePathPrefix is the URL or logical prefix used when exposing this resource via the REST API
	// (e.g., "cronjobs" for "/api/v1/cronjobs").
	ResourcePathPrefix string
}

type ServerGen struct {
	// Computed/derived fields (not serialized).
	Proj              *Project `yaml:"-"`
	Name              string   `yaml:"-"`
	EnvironmentPrefix string   `yaml:"-"`
	// APIImportPath is like: v1 "module/pkg/api/apiserver/v1"
	APIImportPath   string   `yaml:"-"`
	TypedClientName string   `yaml:"-"`
	R               *RESTGen `yaml:"-"`
}

func NewServerGen(proj *Project, binaryName string) ServerGen {
	s := ServerGen{
		Proj: proj,
		Name: GetComponentName(binaryName),
	}

	// Environment variable prefix: PROJECT_COMPONENT (uppercased).
	s.EnvironmentPrefix = strings.ReplaceAll(fmt.Sprintf("%s_%s",
		strings.ToUpper(proj.M.ProjectName),
		strings.ToUpper(s.Name),
	), "-", "_")

	// Import alias path for API package.
	s.APIImportPath = fmt.Sprintf(`%s "%s/pkg/api/%s/%s"`,
		s.Proj.M.APIAlias,
		s.Proj.M.ModuleName,
		s.Name,
		s.Proj.M.APIVersion,
	)
	return s
}

// TemplateData is the rendering context for templates (project + one component).
type TemplateData struct {
	*Project
	Web *WebServer
	MQ  *MQServer
	Job *JobServer
	CLI *CLITool
}

func (td *TemplateData) AbsPath(relPath string) string {
	return td.Project.Join(relPath)
}

// GetComponentName extracts the component name from a binary name.
func GetComponentName(binaryName string) string {
	parts := strings.Split(binaryName, "-")
	if len(parts) == 2 {
		return parts[1]
	}
	return binaryName
}
