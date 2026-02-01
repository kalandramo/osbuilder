package create

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/enescakir/emoji"
	"github.com/fatih/color"
	stringsutil "github.com/onexstack/onexstack/pkg/util/strings"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/kubectl/pkg/util/templates"

	cmdutil "github.com/onexstack/osbuilder/internal/osbuilder/cmd/util"
	"github.com/onexstack/osbuilder/internal/osbuilder/file"
	"github.com/onexstack/osbuilder/internal/osbuilder/helper"
	"github.com/onexstack/osbuilder/internal/osbuilder/known"
	"github.com/onexstack/osbuilder/internal/osbuilder/types"
	"github.com/onexstack/osbuilder/internal/osbuilder/validation"
)

// ProjectOptions contains flags and runtime state for the 'create project' command.
type ProjectOptions struct {
	RootDir string
	Config  string
	// Hidden flag for base64 encoded project config
	ConfigBase64 string
	ShowTips     bool // Print getting-started hints

	Project *types.Project

	genericiooptions.IOStreams
}

var (
	projectLongDesc = templates.LongDesc(`
Create a new project from a configuration file.

This command scaffolds a project structure, generates boilerplate files,
and prepares build and run artifacts.`)

	projectExamples = templates.Examples(`
  # Generate a project using a config file
  osbuilder create project ./my-project --config ./onexstack.yaml

  # Generate a project in the current directory with default config path
  osbuilder create project .`)
)

// NewProjectOptions returns a new ProjectOptions with default IO streams.
func NewProjectOptions(ioStreams genericiooptions.IOStreams) *ProjectOptions {
	return &ProjectOptions{IOStreams: ioStreams, ShowTips: true}
}

// NewCmdProject returns the 'create project' command definition.
func NewCmdProject(f cmdutil.Factory, ioStreams genericiooptions.IOStreams) *cobra.Command {
	o := NewProjectOptions(ioStreams)

	cmd := &cobra.Command{
		Use:                   "project [DIR]",
		Aliases:               []string{"proj"},
		Short:                 "Create a new project from a config file",
		Long:                  projectLongDesc,
		Example:               projectExamples,
		DisableFlagsInUseLine: true,
		TraverseChildren:      true,
		SilenceUsage:          true,
		SilenceErrors:         true,
		// DIR is optional: defaults to current working directory.
		Args: cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(f, cmd, args))
			cmdutil.CheckErr(o.Validate(cmd, args))
			cmdutil.CheckErr(o.Run(f, o.IOStreams, args))
		},
	}

	cmd.Flags().StringVarP(&o.Config, "config", "c", o.Config, "Path to project config file (default: ./onexstack.yaml under the chosen directory)")

	// Add hidden flags
	cmd.Flags().StringVar(&o.ConfigBase64, "config-base64", "", "Base64 encoded project configuration (hidden flag)")
	cmd.Flags().BoolVar(&o.ShowTips, "show-tips", o.ShowTips, "Print post-run tips.")
	_ = cmd.Flags().MarkHidden("show-tips")
	_ = cmd.Flags().MarkHidden("config-base64")

	return cmd
}

// Complete resolves input arguments and loads project metadata.
func (o *ProjectOptions) Complete(_ cmdutil.Factory, _ *cobra.Command, args []string) error {
	// Resolve root directory (explicit arg or current working directory)
	o.RootDir, _ = os.Getwd()
	if len(args) == 1 {
		abs, err := filepath.Abs(args[0])
		if err != nil {
			return fmt.Errorf("resolve directory: %w", err)
		}
		o.RootDir = abs
	}

	// Default config path if not provided (after RootDir is known)
	if strings.TrimSpace(o.Config) == "" {
		o.Config = filepath.Join(o.RootDir, known.ProjectFileName)
	}

	var proj *types.Project
	var err error

	// Check if base64 config is provided (priority over file config)
	if o.ConfigBase64 != "" {
		proj, err = LoadProjectFromBase64(o.ConfigBase64)
	} else {
		// Load project configuration from file
		proj, err = LoadProjectFromFile(o.Config)
	}
	if err != nil {
		return fmt.Errorf("failed to read project configuration: %w", err)
	}

	// Fill generated data
	proj.M = (&types.ProjectGen{
		WorkDir:    o.RootDir,
		APIVersion: "v1",
		APIAlias:   "v1",
		ModuleName: MustModulePath(proj.Metadata.ModulePath, o.RootDir),
	}).Complete()

	proj.M.ProjectName = filepath.Base(o.RootDir)
	proj.M.RegistryPrefix = proj.Metadata.Image.RegistryPrefix

	o.Project = correctProjectConfig(proj)

	// Generate per-webserver files
	for i, ws := range o.Project.WebServers {
		o.Project.WebServers[i] = ws.Complete(o.Project)
	}

	// Generate per-mqserver files
	for i, mq := range o.Project.MQServers {
		o.Project.MQServers[i] = mq.Complete(o.Project)
	}

	// Generate per-jobserver files
	for i, job := range o.Project.JobServers {
		o.Project.JobServers[i] = job.Complete(o.Project)
	}

	// Generate per-clitool files
	for i, cli := range o.Project.CLITools {
		o.Project.CLITools[i] = cli.Complete(o.Project)
	}

	return nil
}

// Validate ensures options and loaded project are consistent.
func (o *ProjectOptions) Validate(_ *cobra.Command, _ []string) error {
	if o.Project == nil {
		return fmt.Errorf("project is not loaded")
	}

	// Validate module path format
	if err := validation.ValidateModulePath(o.Project.M.ModuleName); err != nil {
		return fmt.Errorf("invalid module path %q: %w", o.Project.M.ModuleName, err)
	}

	if err := validateMetadata(o.Project.Metadata); err != nil {
		return err
	}

	// Validate web frameworks and storage types (per web server)
	for _, ws := range o.Project.WebServers {
		// Web framework
		wf := strings.TrimSpace(ws.WebFramework)
		if !known.AvailableWebFrameworks.Has(wf) {
			return fmt.Errorf(
				"web server %q: unsupported webFramework %q; supported: %s",
				ws.Name, wf, strings.Join(known.AvailableWebFrameworks.UnsortedList(), ", "),
			)
		}

		// Storage type
		st := strings.TrimSpace(ws.StorageType)
		if !known.AvailableStorageTypes.Has(st) {
			return fmt.Errorf(
				"web server %q: unsupported storageType %q; supported: %s",
				ws.Name, st, strings.Join(known.AvailableStorageTypes.UnsortedList(), ", "),
			)
		}

		// Service registry type
		sr := strings.TrimSpace(ws.ServiceRegistry)
		if !known.AvailableServiceRegistry.Has(sr) {
			return fmt.Errorf(
				"web server %q: unsupported serviceRegistry %q; supported: %s",
				ws.Name, st, strings.Join(known.AvailableServiceRegistry.UnsortedList(), ", "),
			)
		}
	}

	// Validate mq frameworks and storage types (per mq server)
	for _, mq := range o.Project.MQServers {
		// MQ framework
		if !known.AvailableMQFrameworks.Has(mq.MQFramework) {
			return fmt.Errorf(
				"mq server %q: unsupported MQFramework %q; supported: %s",
				mq.Name, mq.MQFramework, strings.Join(known.AvailableMQFrameworks.UnsortedList(), ", "),
			)
		}

		// Storage type
		st := strings.TrimSpace(mq.StorageType)
		if !known.AvailableStorageTypes.Has(st) {
			return fmt.Errorf(
				"web server %q: unsupported storageType %q; supported: %s",
				mq.Name, st, strings.Join(known.AvailableStorageTypes.UnsortedList(), ", "),
			)
		}
	}

	// Validate storage types.
	for _, job := range o.Project.JobServers {
		// Storage type
		st := strings.TrimSpace(job.StorageType)
		if !known.AvailableStorageTypes.Has(st) {
			return fmt.Errorf(
				"web server %q: unsupported storageType %q; supported: %s",
				job.Name, st, strings.Join(known.AvailableStorageTypes.UnsortedList(), ", "),
			)
		}
	}

	for range o.Project.CLITools {
	}

	return nil
}

// Run generates the project files and prints next steps.
func (o *ProjectOptions) Run(f cmdutil.Factory, ioStreams genericiooptions.IOStreams, _ []string) (err error) {
	defer func() { helper.RecordOSBuilderUsage("project", err) }()

	fm := file.NewFileManager(o.RootDir, false)

	if err := o.Generate(f, fm); err != nil {
		return err
	}

	// Copy third-party dependencies
	fs := helper.NewFileSystem("/")
	if err := fm.CopyFiles(filepath.Join(fs.BasePath, "/project/third_party"), filepath.Join(o.RootDir, "third_party")); err != nil {
		return fmt.Errorf("copy third_party: %w", err)
	}

	if o.ShowTips {
		o.PrintGettingStarted()
	}
	return nil
}

// PrintGettingStarted prints follow-up commands to bootstrap the project.
func (o *ProjectOptions) PrintGettingStarted() {
	baseName := filepath.Base(o.Project.M.WorkDir)

	fmt.Printf("\n%s Project creation succeeded %s\n", emoji.CheckMarkButton, color.GreenString(baseName))
	fmt.Printf("%s Use the following command to start the project %s:\n\n", emoji.Parse(":computer:"), emoji.Parse(":point_down:"))

	fmt.Println(
		color.WhiteString("$ cd %s", o.Project.M.WorkDir),
		color.CyanString("# enter project directory"),
	)

	if o.Project.Metadata.MakefileMode != known.MakefileModeNone {
		fmt.Println(
			color.WhiteString("$ make deps"),
			color.CyanString("# (Optional, executed when dependencies missing) Install tools required by project."),
		)

		switch n := len(o.Project.MQServers); {
		case n == 1:
			fmt.Println(
				color.WhiteString("$ make protoc.%s", o.Project.MQServers[0].Name),
				color.CyanString("# generate gRPC code"),
			)
		case n > 0:
			fmt.Println(
				color.WhiteString("$ make protoc"),
				color.CyanString("# generate gRPC code"),
			)
		}
	}

	fmt.Println(
		color.WhiteString("$ go get cloud.google.com/go/compute@latest"),
		color.CyanString("# to resolve `cloud.google.com/go/compute/metadata: ambiguous import`"),
	)
	fmt.Println(
		color.WhiteString("$ go get cloud.google.com/go/compute/metadata@latest"),
	)
	fmt.Println(
		color.WhiteString("$ go mod tidy"),
		color.CyanString("# tidy dependencies"),
	)
	fmt.Println(
		color.WhiteString("$ go generate ./..."),
		color.CyanString("# run all go:generate directives"),
	)

	if o.Project.Metadata.MakefileMode == known.MakefileModeNone {
		PrintClosingTips(o.Project.M.ProjectName)
		return
	}

	if len(o.Project.WebServers) == 1 {
		ws := o.Project.WebServers[0]
		if o.Project.Metadata.MakefileMode != known.MakefileModeNone {
			fmt.Println(
				color.WhiteString("$ make build BINS=%s", ws.BinaryName),
				color.CyanString("# build %s", ws.BinaryName),
			)
		}
		if ws.StorageType == known.StorageTypeMemory {
			fmt.Println(
				color.WhiteString("$ _output/platforms/%s/%s/%s", runtime.GOOS, runtime.GOARCH, ws.BinaryName),
				color.CyanString("# run the compiled server"),
			)
			if ws.WebFramework == known.WebFrameworkGin && ws.WithHealthz {
				fmt.Println(
					color.WhiteString("$ curl http://127.0.0.1:5555/healthz"),
					color.CyanString("# test with the health endpoint"),
				)
			}
			if ws.WebFramework == known.WebFrameworkGRPC {
				if ws.WithUser {
					fmt.Println(
						color.WhiteString("$ go run examples/client/user/main.go"),
						color.CyanString("# run user client to test the API"),
					)
				} else if ws.WithHealthz {
					fmt.Println(
						color.WhiteString("$ go run examples/client/health/main.go"),
						color.CyanString("# run health client to test the API"),
					)
				}
			}
		}
	} else {
		fmt.Println(
			color.WhiteString("$ make build"),
			color.CyanString("# Build all available binaries"),
		)
	}

	PrintClosingTips(o.Project.M.ProjectName)
}

// Generate creates project-level files and per-component code based on templates.
func (o *ProjectOptions) Generate(f cmdutil.Factory, fm *file.FileManager) error {
	// Persist project config into the root directory
	if err := o.Project.Save(o.Project.Join(known.ProjectFileName)); err != nil {
		return err
	}

	funcs := helper.GetTemplateFuncMap()

	// Project-level templates
	projectFiles := map[string]string{
		filepath.Join("go.mod"):                                     "/project/go.mod",
		filepath.Join("README.md"):                                  "/project/README.md",
		filepath.Join("scripts/boilerplate.txt"):                    "/project/scripts/boilerplate.txt",
		filepath.Join(".gitignore"):                                 "/project/gitignore.tpl",
		filepath.Join(".golangci.yaml"):                             "/project/golangci.yaml",
		filepath.Join(".protolint.yaml"):                            "/project/protolint.yaml",
		filepath.Join("docs/images/.keep"):                          "/keep.tpl",
		filepath.Join("docs/devel/en-US/.keep"):                     "/keep.tpl",
		filepath.Join("docs/devel/zh-CN/.keep"):                     "/keep.tpl",
		filepath.Join("docs/guide/en-US/.keep"):                     "/keep.tpl",
		filepath.Join("docs/guide/zh-CN/README.md"):                 "/project/docs/guide/zh-CN/README.md",
		filepath.Join("docs/guide/zh-CN/announcements.md"):          "/project/docs/guide/zh-CN/announcements.md",
		filepath.Join("docs/guide/zh-CN/introduction/README.md"):    "/project/docs/guide/zh-CN/introduction/README.md",
		filepath.Join("docs/guide/zh-CN/quickstart/README.md"):      "/project/docs/guide/zh-CN/quickstart/README.md",
		filepath.Join("docs/guide/zh-CN/installation/README.md"):    "/project/docs/guide/zh-CN/installation/README.md",
		filepath.Join("docs/guide/zh-CN/operation-guide/README.md"): "/project/docs/guide/zh-CN/operation-guide/README.md",
		filepath.Join("docs/guide/zh-CN/best-practice/README.md"):   "/project/docs/guide/zh-CN/best-practice/README.md",
		filepath.Join("docs/guide/zh-CN/faq/README.md"):             "/project/docs/guide/zh-CN/faq/README.md",
		// Scripts
		filepath.Join("scripts/coverage.awk"): "/project/scripts/coverage.awk",
	}

	// Deployment-specific files
	switch o.Project.Metadata.DeploymentMethod {
	case known.DeploymentModeSystemd:
		projectFiles[filepath.Join("init", "README.md")] = "/project/init/README.md"
	}

	// Makefile
	switch o.Project.Metadata.MakefileMode {
	case known.MakefileModeUnstructured:
		projectFiles["Makefile"] = "/project/Makefile.unstructed"
	case known.MakefileModeStructured:
		projectFiles["Makefile"] = "/project/Makefile.structed"
		projectFiles["scripts/make-rules/all.mk"] = "/project/scripts/make-rules/all.mk"
		projectFiles["scripts/make-rules/common.mk"] = "/project/scripts/make-rules/common.mk"
		projectFiles["scripts/make-rules/generate.mk"] = "/project/scripts/make-rules/generate.mk"
		projectFiles["scripts/make-rules/golang.mk"] = "/project/scripts/make-rules/golang.mk"
		projectFiles["scripts/make-rules/tools.mk"] = "/project/scripts/make-rules/tools.mk"
		projectFiles["scripts/make-rules/image.mk"] = "/project/scripts/make-rules/image.mk"
	default:
	}

	// Generate project-level files
	if err := helper.RenderTemplate(fm, projectFiles, funcs, &types.TemplateData{Project: o.Project}); err != nil {
		return err
	}

	// Generate per-webserver files
	for _, ws := range o.Project.WebServers {
		data := types.TemplateData{Project: o.Project, Web: ws}

		// 生成web server主文件
		if err := helper.RenderTemplate(fm, ws.Pairs(), funcs, &data); err != nil {
			return err
		}

		// 生成 client 类型的 fake 文件
		for _, kind := range ws.Clients {
			ws.TypedClientName = helper.ToLower(kind)
			tplFile := "/project/internal/apiserver/pkg/clientset/typed/fake/fake.go"
			pairs := map[string]string{
				filepath.Join(ws.PkgDir(), fmt.Sprintf("clientset/typed/%s/%s.go", ws.TypedClientName, ws.TypedClientName)): tplFile,
			}
			if err := helper.RenderTemplate(fm, pairs, nil, &data); err != nil {
				return err
			}
		}
	}

	// Generate per-mqserver files
	for _, mq := range o.Project.MQServers {
		data := types.TemplateData{Project: o.Project, MQ: mq}

		// 生成web server主文件
		if err := helper.RenderTemplate(fm, mq.Pairs(), funcs, &data); err != nil {
			return err
		}

		// 生成 client 类型的 fake 文件
		for _, kind := range mq.Clients {
			mq.TypedClientName = helper.ToLower(kind)
			tplFile := "/project/internal/mqserver/pkg/clientset/typed/fake/fake.go"
			pairs := map[string]string{
				filepath.Join(mq.PkgDir(), fmt.Sprintf("clientset/typed/%s/%s.go", mq.TypedClientName, mq.TypedClientName)): tplFile,
			}
			if err := helper.RenderTemplate(fm, pairs, nil, &data); err != nil {
				return err
			}
		}
	}

	// Generate per-jobserver files
	for _, job := range o.Project.JobServers {
		data := types.TemplateData{Project: o.Project, Job: job}

		// 生成web server主文件
		if err := helper.RenderTemplate(fm, job.Pairs(), funcs, &data); err != nil {
			return err
		}

		// 生成 client 类型的 fake 文件
		for _, kind := range job.Clients {
			job.TypedClientName = helper.ToLower(kind)
			tplFile := "/project/internal/jobserver/pkg/clientset/typed/fake/fake.go"
			pairs := map[string]string{
				filepath.Join(job.PkgDir(), fmt.Sprintf("clientset/typed/%s/%s.go", job.TypedClientName, job.TypedClientName)): tplFile,
			}
			if err := helper.RenderTemplate(fm, pairs, nil, &data); err != nil {
				return err
			}
		}
	}

	// Generate per-clitool files
	for _, cli := range o.Project.CLITools {
		data := types.TemplateData{Project: o.Project, CLI: cli}

		// 生成web server主文件
		if err := helper.RenderTemplate(fm, cli.Pairs(), funcs, &data); err != nil {
			return err
		}

		// 生成 client 类型的 fake 文件
		for _, kind := range cli.Clients {
			cli.TypedClientName = helper.ToLower(kind)
			tplFile := "/project/internal/mbctl/pkg/clientset/typed/fake/fake.go"
			pairs := map[string]string{
				filepath.Join(cli.PkgDir(), fmt.Sprintf("clientset/typed/%s/%s.go", cli.TypedClientName, cli.TypedClientName)): tplFile,
			}
			if err := helper.RenderTemplate(fm, pairs, nil, &data); err != nil {
				return err
			}
		}
	}

	return nil
}

// correctProjectConfig fixes inconsistent project configuration.
func correctProjectConfig(proj *types.Project) *types.Project {
	if proj.Metadata.DeploymentMethod == "" {
		proj.Metadata.DeploymentMethod = known.DeploymentModeDocker
	}
	if proj.Metadata.MakefileMode == "" {
		proj.Metadata.MakefileMode = known.MakefileModeUnstructured
	}
	if proj.Metadata.ShortDescription == "" {
		proj.Metadata.ShortDescription = "TODO: Update the short description of the binary file."
	}
	if proj.Metadata.LongMessage == "" {
		proj.Metadata.LongMessage = "TODO: Update the detailed description of the binary file."
	}
	if proj.Metadata.Image.RegistryPrefix == "" {
		proj.Metadata.Image.RegistryPrefix = "docker.io/_undefined"
	}
	if proj.Metadata.Image.DockerfileMode == "" {
		proj.Metadata.Image.DockerfileMode = known.DockerfileModeCombined
	}

	for _, ws := range proj.WebServers {
		if ws.WebFramework == "" {
			ws.WebFramework = known.WebFrameworkGin
		}
		if ws.StorageType == "" {
			ws.StorageType = known.StorageTypeMemory
		}
		if ws.StorageType == known.StorageTypeMySQL {
			ws.StorageType = known.StorageTypeMariaDB
		}
		if ws.WebFramework != known.WebFrameworkGRPC && ws.WebFramework != known.WebFrameworkGRPCGateway {
			ws.GRPCServiceName = ""
		}
		if ws.ServiceRegistry == "" {
			ws.ServiceRegistry = known.ServiceRegistryNone
		}
		if ws.WebFramework != known.WebFrameworkGin {
			ws.WithWS = false
		}
	}

	for _, mq := range proj.MQServers {
		if mq.StorageType == "" {
			mq.StorageType = known.StorageTypeMemory
		}
		if mq.StorageType == known.StorageTypeMySQL {
			mq.StorageType = known.StorageTypeMariaDB
		}
	}

	for _, job := range proj.JobServers {
		if job.StorageType == "" {
			job.StorageType = known.StorageTypeMemory
		}
		if job.StorageType == known.StorageTypeMySQL {
			job.StorageType = known.StorageTypeMariaDB
		}
	}

	return proj
}

// validateMetadata validate metadata.
func validateMetadata(md *types.Metadata) error {
	// Validate deployment mode (project-level)
	dep := strings.TrimSpace(md.DeploymentMethod)
	if !known.AvailableDeploymentModes.Has(dep) {
		return fmt.Errorf(
			"unsupported metadata.deploymentMode %q; supported: %s",
			dep, strings.Join(known.AvailableDeploymentModes.UnsortedList(), ", "),
		)
	}

	// If use cloud-native deploy method, need to generate Dockerfile.
	if stringsutil.StringIn(md.DeploymentMethod, []string{known.DeploymentModeDocker, known.DeploymentModeKubernetes}) {
		if err := validateImageConfig(md.Image); err != nil {
			return err
		}
	}

	// Validate makefile mode (project-level)
	if !known.AvailableMakefileModes.Has(md.MakefileMode) {
		return fmt.Errorf(
			"unsupported metadata.makefileMode %q; supported: %s",
			md.MakefileMode,
			strings.Join(known.AvailableMakefileModes.UnsortedList(), ", "),
		)
	}

	if md.MakefileMode == known.MakefileModeNone {
		fmt.Println(color.YellowString("Warning! If `makefileMode` is set to none, you must manually execute commands to build the source code."))
	}

	return nil
}

func validateImageConfig(image types.ImageConfig) error {
	if image.RegistryPrefix == "" {
		return fmt.Errorf("metadata.image.registryPrefix cannot be empty")
	}

	// Validate dockerfile mode (project-level)
	if !known.AvailableDockerfileModes.Has(image.DockerfileMode) {
		return fmt.Errorf(
			"unsupported metadata.image.dockerfileMode %q; supported: %s",
			image.DockerfileMode, strings.Join(known.AvailableDockerfileModes.UnsortedList(), ", "),
		)
	}

	if !known.AvailableDistrolessModes.Has(image.DistrolessMode) {
		return fmt.Errorf(
			"unsupported metadata.image.distrolessMode %q; supported: %s",
			image.DistrolessMode, strings.Join(known.AvailableDistrolessModes.UnsortedList(), ", "),
		)
	}

	return nil
}
