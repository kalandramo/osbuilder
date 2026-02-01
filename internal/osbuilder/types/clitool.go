package types

import (
	"path/filepath"
)

// CLITool describes a web server component to generate (HTTP/gRPC/etc).
type CLITool struct {
	// BinaryName is the CLI binary name (e.g., "mb-apiserver").
	BinaryName string   `yaml:"binaryName"`
	Clients    []string `yaml:"clients,omitempty"`

	// Computed/derived fields (not serialized).
	ServerGen `yaml:"-"`
}

// Complete populates derived fields and sensible defaults.
func (cli *CLITool) Complete(proj *Project) *CLITool {
	cli.ServerGen = NewServerGen(proj, cli.BinaryName)
	return cli
}

// BaseDir returns the component base directory: internal/<component>.
func (cli *CLITool) BaseDir() string {
	return filepath.Join("internal", cli.Name)
}

// CMDDir returns the command directory for the component.
func (cli *CLITool) CMDDir() string {
	return filepath.Join(cli.BaseDir(), "cmd")
}

func (cli *CLITool) PkgDir() string {
	return filepath.Join(cli.BaseDir(), "pkg")
}

// PrepareRESTMetadata constructs REST metadata for a given kind.
func (cli *CLITool) PrepareRESTMetadata(cmdPath string) {
	cli.R = prepareRESTMetadata(cli.Proj.M.APIVersion, cmdPath)
}

// Pairs returns a map of destination relative paths to template paths.
// It drives file generation for this component.
func (cli *CLITool) Pairs() map[string]string {
	// Local shortcuts to reduce repetition.
	baseDir := cli.BaseDir()
	cmdDir := cli.CMDDir()
	pkgDir := cli.PkgDir()

	pairs := map[string]string{}
	add := func(dst, tpl string) {
		pairs[dst] = tpl
	}

	// Common command and component scaffolding.
	add(filepath.Join("cmd", cli.BinaryName, "main.go"), "/project/cmd/mbctl/main.go")
	add(filepath.Join(cli.Proj.Configs(), cli.BinaryName+".yaml"), "/project/configs/mbctl.yaml")
	add(filepath.Join(baseDir, "doc.go"), "/project/internal/mbctl/doc.go")

	add(filepath.Join(baseDir, "util/helper.go"), "/project/internal/mbctl/util/helper.go")
	add(filepath.Join(baseDir, "util/interrupt/interrupt.go"), "/project/internal/mbctl/util/interrupt/interrupt.go")
	add(filepath.Join(baseDir, "util/options/options.go"), "/project/internal/mbctl/util/options/options.go")
	add(filepath.Join(baseDir, "util/options/server_options.go"), "/project/internal/mbctl/util/options/server_options.go")
	add(filepath.Join(baseDir, "util/options/user_options.go"), "/project/internal/mbctl/util/options/user_options.go")
	add(filepath.Join(baseDir, "util/templates/templates.go"), "/project/internal/mbctl/util/templates/templates.go")

	add(filepath.Join(cmdDir, "alpha.go"), "/project/internal/mbctl/cmd/alpha.go")
	add(filepath.Join(cmdDir, "cmd.go"), "/project/internal/mbctl/cmd/cmd.go")
	add(filepath.Join(cmdDir, "profiling.go"), "/project/internal/mbctl/cmd/profiling.go")

	add(filepath.Join(cmdDir, "options/options.go"), "/project/internal/mbctl/cmd/options/options.go")
	add(filepath.Join(cmdDir, "version/skew_warning.go"), "/project/internal/mbctl/cmd/version/skew_warning.go")
	add(filepath.Join(cmdDir, "version/version.go"), "/project/internal/mbctl/cmd/version/version.go")
	add(filepath.Join(cmdDir, "fake/fake.go"), "/project/internal/mbctl/cmd/fake/fake.go")
	add(filepath.Join(cmdDir, "all/all.go"), "/project/internal/mbctl/cmd/all/all.go")

	add(filepath.Join(cmdDir, "util/factory.go"), "/project/internal/mbctl/cmd/util/factory.go")
	add(filepath.Join(cmdDir, "util/helper.go"), "/project/internal/mbctl/cmd/util/helper.go")
	add(filepath.Join(cmdDir, "util/registry.go"), "/project/internal/mbctl/cmd/util/registry.go")
	add(filepath.Join(cmdDir, "util/wire.go"), "/project/internal/mbctl/cmd/util/wire.go")
	add(filepath.Join(cmdDir, "util/wire_gen.go"), "/project/internal/mbctl/cmd/util/wire_gen.go")

	if len(cli.Clients) > 0 {
		add(filepath.Join(pkgDir, "clientset/clientset.go"), "/project/internal/mbctl/pkg/clientset/clientset.go")
	}

	return pairs
}
