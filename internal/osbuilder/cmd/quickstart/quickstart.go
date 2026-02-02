package quickstart

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/kubectl/pkg/util/templates"

	"github.com/onexstack/osbuilder/internal/osbuilder/cmd/create"
	cmdutil "github.com/onexstack/osbuilder/internal/osbuilder/cmd/util"
	"github.com/onexstack/osbuilder/internal/osbuilder/helper"
	"github.com/onexstack/osbuilder/internal/osbuilder/types"
)

// QuickstartOptions holds flags and runtime context for the 'create quickstart' command.
type QuickstartOptions struct {
	ProjectRootDir string

	// Project configuration parameters
	ModuleName  string // Go module name
	ProjectName string // Project name

	Author         string // Project author
	Email          string // Author email
	MakefileMode   string // Makefile mode
	DeploymentMode string // Deployment mode
	RegistryPrefix string // Container registry prefix
	DockerfileMode string
	DistrolessMode string // Use distroless base image

	AppPrefix       string   // eg: mb
	Kinds           []string // Resource kinds to generate (snake_case recommended)
	WebFramework    string   // Web framework to use
	WithUser        bool     // Include user management logic
	WithOtel        bool     // Enable OpenTelemetry
	WithWS          bool     // Enable websocket
	WithPreloader   bool     // Enable data pre-load
	Clients         []string
	ServiceRegistry string // Service registry type

	genericiooptions.IOStreams
}

var (
	quickstartLongDesc = templates.LongDesc(`
		Create a quickstart demo project with sensible defaults.

		This command scaffolds a complete project structure with configurable options
		for web framework, deployment mode, and other common settings. It's designed
		to help you get started quickly with a working project template.`)

	quickstartExamples = templates.Examples(`
		# Create a quickstart project with default settings
		osbuilder quickstart ./my-demo

		# Create with custom module name and web framework
		osbuilder quickstart ./my-project \
		  --module-name github.com/myorg/myproject \
		  --web-framework grpc

		# Create with full customization
		osbuilder quickstart ./enterprise-api \
		  --module-name github.com/company/enterprise-api \
		  --project-name enterprise-api \
		  --author "孔令飞" \
		  --email colin404@foxmail.com \
		  --makefile-mode structured \
		  --deployment-mode kubernetes \
		  --registry-prefix docker.io/company \
		  --distroless-mode auto \
		  --app-prefix mb \
		  --kinds cron_job,job \
		  --web-framework gin \
		  --with-user \
		  --with-otel \
		  --service-registry polaris`)
)

// NewQuickstartOptions creates a default QuickstartOptions.
func NewQuickstartOptions(io genericiooptions.IOStreams) *QuickstartOptions {
	return &QuickstartOptions{
		ModuleName:     "",
		ProjectName:    "miniblog",
		Author:         "孔令飞",
		Email:          "colin404@foxmail.com",
		MakefileMode:   "unstructured",
		DeploymentMode: "kubernetes",
		RegistryPrefix: "", // Will be set to docker.io/<project-name> if empty
		DockerfileMode: "combined",
		DistrolessMode: "auto",
		AppPrefix:      "mb",
		Kinds: []string{
			"post", "comment", "tag", "follow", "follower", "friend",
			"block", "like", "bookmark", "share", "report", "vote",
		},
		WebFramework:    "gin",
		WithUser:        false,
		WithOtel:        true,
		WithWS:          true,
		WithPreloader:   true,
		Clients:         []string{},
		ServiceRegistry: "none",
		IOStreams:       io,
	}
}

// NewCmdQuickstart builds the 'create quickstart' cobra command.
func NewCmdQuickstart(factory cmdutil.Factory, ioStreams genericiooptions.IOStreams) *cobra.Command {
	o := NewQuickstartOptions(ioStreams)

	cmd := &cobra.Command{
		Use:                   "quickstart [PROJECT_NAME]",
		DisableFlagsInUseLine: true,
		Short:                 "Create a quickstart demo project",
		Long:                  quickstartLongDesc,
		Example:               quickstartExamples,
		SilenceUsage:          true,
		SilenceErrors:         true,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("requires exactly one argument (project name), but received %d", len(args))
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(factory, cmd, args))
			cmdutil.CheckErr(o.Validate(cmd, args))
			cmdutil.CheckErr(o.Run(factory, ioStreams, args))
		},
	}

	// Project configuration flags
	cmd.Flags().StringVar(&o.ModuleName, "module-name", o.ModuleName, "Go module name for the project. Default: onexstack/onexstack/<PROJECT_NAME>")
	// cmd.Flags().StringVar(&o.ProjectName, "project-name", o.ProjectName, "Project name")
	cmd.Flags().StringVar(&o.Author, "author", o.Author, "Project author name")
	cmd.Flags().StringVar(&o.Email, "email", o.Email, "Author email address")
	cmd.Flags().StringVar(&o.MakefileMode, "makefile-mode", o.MakefileMode, "Makefile mode (none, unstructured, structured)")
	cmd.Flags().StringVar(&o.DeploymentMode, "deployment-mode", o.DeploymentMode, "Deployment mode (docker, kubernetes, systemd)")
	cmd.Flags().StringVar(&o.RegistryPrefix, "registry-prefix", o.RegistryPrefix, "Container registry prefix (default: docker.io/<project-name>)")
	cmd.Flags().StringVar(&o.DockerfileMode, "dockerfile-mode", o.DockerfileMode, "Specify the dockerfile mode(options: none, runtime-only, multi-stage, combined)")
	cmd.Flags().StringVar(&o.DistrolessMode, "distroless-mode", o.DistrolessMode, "Base image selection mode (options: always, never, auto)")
	cmd.Flags().StringVar(&o.AppPrefix, "app-prefix", o.AppPrefix, "The common prefix for application binaries (e.g., set 'mb' for 'mb-apiserver').")
	cmd.Flags().StringSliceVarP(&o.Kinds, "kinds", "", o.Kinds, "Resource kinds to generate in snake_case (e.g., cron_job).")
	cmd.Flags().StringVar(&o.WebFramework, "web-framework", o.WebFramework, "Web framework to use (gin, grpc, grpc-gateway)")
	cmd.Flags().BoolVar(&o.WithUser, "with-user", o.WithUser, "Include user management, authentication and authorization logic")
	cmd.Flags().BoolVar(&o.WithOtel, "with-otel", o.WithOtel, "Enable OpenTelemetry support")
	cmd.Flags().BoolVar(&o.WithWS, "with-ws", o.WithWS, "Enable websocket support")
	cmd.Flags().BoolVar(&o.WithPreloader, "with-preloader", o.WithPreloader, "Enable data preload feature.")
	cmd.Flags().StringSliceVar(&o.Clients, "clients", o.Clients, "Define clientset.")
	cmd.Flags().StringVar(&o.ServiceRegistry, "service-registry", o.ServiceRegistry, "Service registry type (none, etcd, consul)")

	return cmd
}

// Complete resolves working directory and builds project configuration.
func (o *QuickstartOptions) Complete(factory cmdutil.Factory, cmd *cobra.Command, args []string) error {
	if o.ModuleName == "" {
		o.ModuleName = "github.com/onexstack/" + o.ProjectName
	}

	// Set default registry prefix if not provided
	if o.RegistryPrefix == "" {
		o.RegistryPrefix = fmt.Sprintf("docker.io/%s", o.ProjectName)
	}

	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	o.ProjectName = args[0]
	o.ProjectRootDir = filepath.Join(wd, o.ProjectName)
	return nil
}

// Validate checks required inputs and validates configuration options.
func (o *QuickstartOptions) Validate(cmd *cobra.Command, args []string) error {
	return nil
}

// Run generates the quickstart project files and prints next steps.
func (o *QuickstartOptions) Run(f cmdutil.Factory, ioStreams genericiooptions.IOStreams, args []string) (err error) {
	defer func() { helper.RecordOSBuilderUsage("project", err) }()

	fmt.Printf("\n🍺 Creating quickstart project %s...\n", color.GreenString(o.ProjectName))
	projectString := helper.NewFileSystem("/").Content("/project.yaml")
	if projectString == "" {
		return fmt.Errorf("project template not found")
	}

	// 创建 reader 并解析 YAML
	reader := strings.NewReader(projectString)
	proj, err := create.DecodeProjectYAML(reader, true)
	if err != nil {
		return fmt.Errorf("decode project template: %w", err)
	}

	modifiedProject := o.applyQuickstartOptions(proj)

	yamlString, err := projectToYAMLString(modifiedProject)
	if err != nil {
		return fmt.Errorf("convert project to yaml: %w", err)
	}

	encodedString := base64.StdEncoding.EncodeToString([]byte(yamlString))

	projectCmd := create.NewCmdProject(f, ioStreams)
	projectCmd.SetArgs([]string{o.ProjectName, "--config-base64", encodedString, "--show-tips=false"})

	// 执行命令
	if err := projectCmd.Execute(); err != nil {
		fmt.Printf("Error executing project command: %v\n", err)
	}

	commonArgs := []string{
		"--root-dir", o.ProjectRootDir,
		"--kinds", strings.Join(o.Kinds, ","),
		"--show-tips=false",
	}

	// 添加 API 接口
	apiCmd := create.NewCmdAPI(f, ioStreams)
	apiCmd.SetArgs(commonArgs)
	if err := apiCmd.Execute(); err != nil {
		fmt.Printf("Error executing api command: %v\n", err)
	}

	// 添加异步任务
	jobCmd := create.NewCmdJob(f, ioStreams)
	jobCmd.SetArgs(commonArgs)
	if err := jobCmd.Execute(); err != nil {
		fmt.Printf("Error executing job command: %v\n", err)
	}

	// 添加消息队列
	mqCmd := create.NewCmdMQ(f, ioStreams)
	mqCmd.SetArgs(commonArgs)
	if err := mqCmd.Execute(); err != nil {
		fmt.Printf("Error executing message queue command: %v\n", err)
	}

	cmdCmd := create.NewCmdCmd(f, ioStreams)
	cmdCmd.SetArgs(commonArgs)
	if err := cmdCmd.Execute(); err != nil {
		fmt.Printf("Error executing cmd command: %v\n", err)
	}

	projectOptions := create.NewProjectOptions(ioStreams)
	projectOptions.ConfigBase64 = encodedString
	_ = projectOptions.Complete(f, nil, []string{o.ProjectName})
	projectOptions.PrintGettingStarted()
	return nil
}

// BuildProject constructs a project configuration from command-line options.
func (o *QuickstartOptions) BuildProject() *types.Project {
	return nil
}

// applyQuickstartOptions 根据 quickstart 命令行参数修改项目配置
func (o *QuickstartOptions) applyQuickstartOptions(project *types.Project) *types.Project {
	// 修改 metadata 配置
	project.Metadata.ModulePath = o.ModuleName

	if o.Author != "" {
		project.Metadata.Author = o.Author
	}
	if o.Email != "" {
		project.Metadata.Email = o.Email
	}
	if o.MakefileMode != "" {
		project.Metadata.MakefileMode = o.MakefileMode
	}
	if o.DeploymentMode != "" {
		project.Metadata.DeploymentMethod = o.DeploymentMode
	}
	if o.RegistryPrefix != "" {
		project.Metadata.Image.RegistryPrefix = o.RegistryPrefix
	}
	project.Metadata.Image.DockerfileMode = o.DockerfileMode
	project.Metadata.Image.DistrolessMode = o.DistrolessMode

	// 修改 WebServers 配置
	for i := range project.WebServers {
		if o.AppPrefix != "" {
			project.WebServers[i].BinaryName = fmt.Sprintf("%s-apiserver$d", o.AppPrefix, i)
		}
		if o.WebFramework != "" {
			project.WebServers[i].WebFramework = o.WebFramework
		}
		project.WebServers[i].WithUser = o.WithUser
		project.WebServers[i].WithOTel = o.WithOtel
		project.WebServers[i].WithWS = o.WithWS
		project.WebServers[i].WithPreloader = o.WithPreloader
		project.WebServers[i].Clients = o.Clients
		if o.ServiceRegistry != "" {
			project.WebServers[i].ServiceRegistry = o.ServiceRegistry
		}
	}

	// 修改 WebServers 配置
	for i := range project.WebServers {
		binaryName := project.WebServers[i].BinaryName
		if o.AppPrefix != "" {
			binaryName = o.AppPrefix + "-apiserver"
		}
		if len(project.WebServers) > 1 {
			binaryName = fmt.Sprintf("%s%d", binaryName, i)
		}
		project.WebServers[i].BinaryName = binaryName

		if o.WebFramework != "" {
			project.WebServers[i].WebFramework = o.WebFramework
		}
		project.WebServers[i].WithUser = o.WithUser
		project.WebServers[i].WithOTel = o.WithOtel
		project.WebServers[i].WithWS = o.WithWS
		project.WebServers[i].WithPreloader = o.WithPreloader
		project.WebServers[i].Clients = o.Clients
		if o.ServiceRegistry != "" {
			project.WebServers[i].ServiceRegistry = o.ServiceRegistry
		}
	}

	// 修改 JobServers 配置
	for i := range project.JobServers {
		binaryName := project.JobServers[i].BinaryName
		if o.AppPrefix != "" {
			binaryName = o.AppPrefix + "-jobserver"
		}
		if len(project.JobServers) > 1 {
			binaryName = fmt.Sprintf("%s%d", binaryName, i)
		}
		project.JobServers[i].BinaryName = binaryName

		project.JobServers[i].WithOTel = o.WithOtel
		project.JobServers[i].WithPreloader = o.WithPreloader
		project.JobServers[i].Clients = o.Clients
	}

	// 修改 MQServers 配置
	for i := range project.MQServers {
		binaryName := project.MQServers[i].BinaryName
		if o.AppPrefix != "" {
			binaryName = o.AppPrefix + "-mqserver"
		}
		if len(project.MQServers) > 1 {
			binaryName = fmt.Sprintf("%s%d", binaryName, i)
		}
		project.MQServers[i].BinaryName = binaryName

		project.MQServers[i].WithOTel = o.WithOtel
		project.MQServers[i].WithPreloader = o.WithPreloader
		project.MQServers[i].Clients = o.Clients
	}

	// 修改 CLITools 配置
	for i := range project.CLITools {
		binaryName := project.CLITools[i].BinaryName
		if o.AppPrefix != "" {
			binaryName = o.AppPrefix + "ctl"
		}
		if len(project.CLITools) > 1 {
			binaryName = fmt.Sprintf("%s%d", binaryName, i)
		}
		project.CLITools[i].BinaryName = binaryName

		project.CLITools[i].Clients = o.Clients
	}

	return project
}

// projectToYAMLString 将项目配置转换为 YAML 格式字符串
func projectToYAMLString(project *types.Project) (string, error) {
	if project == nil {
		return "", fmt.Errorf("project is nil")
	}

	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2) // 设置缩进为 2 个空格

	if err := encoder.Encode(project); err != nil {
		return "", fmt.Errorf("encode project to yaml: %w", err)
	}

	if err := encoder.Close(); err != nil {
		return "", fmt.Errorf("close yaml encoder: %w", err)
	}

	return buf.String(), nil
}
