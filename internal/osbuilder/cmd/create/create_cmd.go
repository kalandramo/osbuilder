package create

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/enescakir/emoji"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/kubectl/pkg/util/templates"

	cmdutil "github.com/onexstack/osbuilder/internal/osbuilder/cmd/util"
	"github.com/onexstack/osbuilder/internal/osbuilder/file"
	"github.com/onexstack/osbuilder/internal/osbuilder/helper"
	"github.com/onexstack/osbuilder/internal/osbuilder/known"
	"github.com/onexstack/osbuilder/internal/osbuilder/types"
)

// CmdOptions holds flags and runtime context for the 'create cmd' command.
type CmdOptions struct {
	RootDir string

	Kinds      []string // Command names to generate (kebab-case recommended)
	BinaryName string   // Target CLI binary name
	Force      bool     // Overwrite files if they exist

	// APIVersion is kept for directory structure consistency
	APIVersion string
	ShowTips   bool // Print getting-started hints

	Project *types.Project // Loaded project metadata

	genericiooptions.IOStreams
}

var (
	cmdLongDesc = templates.LongDesc(`
		Create Cobra commands for your CLI project.

		This command scaffolds command files (implementation, flag parsing) for the given command names.`)

	cmdExamples = templates.Examples(`
		# Create a single command for a specific binary
		osbuilder create cmd --commands create-user --binary-name dmsctl

		# Create multiple commands
		osbuilder create cmd --commands init,version --binary-name dmsctl`)
)

// NewCmdOptions creates a default CmdOptions.
func NewCmdOptions(io genericiooptions.IOStreams) *CmdOptions {
	return &CmdOptions{
		APIVersion: "v1",
		ShowTips:   true,
		IOStreams:  io,
	}
}

// NewCmdCmd builds the 'create cmd' cobra command.
func NewCmdCmd(factory cmdutil.Factory, ioStreams genericiooptions.IOStreams) *cobra.Command {
	o := NewCmdOptions(ioStreams)

	cmd := &cobra.Command{
		Use:                   "cmd",
		DisableFlagsInUseLine: true,
		Short:                 "Create a new CLI command",
		Long:                  cmdLongDesc,
		Example:               cmdExamples,
		SilenceUsage:          true,
		SilenceErrors:         true,
		Args:                  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(factory, cmd, args))
			cmdutil.CheckErr(o.Validate(cmd, args))
			cmdutil.CheckErr(o.Run(args))
		},
	}

	// Flags
	cmd.Flags().StringSliceVarP(&o.Kinds, "kinds", "", o.Kinds, "Command names to generate in kebab-case (e.g., get-profile).")
	cmd.Flags().StringVarP(&o.BinaryName, "binary-name", "b", o.BinaryName, "Target CLI binary name (e.g., dmsctl).")
	cmd.Flags().BoolVarP(&o.Force, "force", "f", o.Force, "Force overwriting of existing files.")

	// Add hidden flags
	cmd.Flags().StringVar(&o.RootDir, "root-dir", "", "Override root directory (hidden flag)")
	cmd.Flags().BoolVar(&o.ShowTips, "show-tips", o.ShowTips, "Print post-run tips.")
	_ = cmd.Flags().MarkHidden("root-dir")
	_ = cmd.Flags().MarkHidden("show-tips")

	return cmd
}

// Complete resolves working directory and loads project metadata.
func (o *CmdOptions) Complete(factory cmdutil.Factory, cmd *cobra.Command, args []string) error {
	if o.RootDir == "" {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		o.RootDir = wd
	}

	proj, err := LoadProjectFromFile(filepath.Join(o.RootDir, known.ProjectFileName))
	if err != nil {
		return err
	}

	// Fill generated data
	proj.M = (&types.ProjectGen{
		WorkDir:    o.RootDir,
		APIVersion: "v1",
		APIAlias:   "v1",
		ModuleName: MustModulePath(proj.Metadata.ModulePath, o.RootDir),
	}).Complete()
	proj.M.ProjectName = filepath.Base(o.RootDir)

	// Smart default: If a single CLI server exists and BinaryName not set, default to it.
	// Note: Assuming proj.CLITools exists similar to proj.JobServers
	if o.BinaryName == "" && len(proj.CLITools) == 1 {
		o.BinaryName = proj.CLITools[0].Name
	}

	o.Project = proj
	return nil
}

// Validate checks required inputs and project state.
func (o *CmdOptions) Validate(cmd *cobra.Command, args []string) error {
	if o.Project == nil {
		return fmt.Errorf("project not loaded")
	}
	if len(o.Kinds) == 0 {
		return fmt.Errorf("at least one command name must be provided via --commands")
	}

	// Validate binary existence
	// Assuming FindCLITool exists in your types package
	cli := o.Project.FindCLITool(o.BinaryName)
	if cli == nil {
		return fmt.Errorf("CLI binary %q not found in project; use --binary-name", o.BinaryName)
	}
	return nil
}

// Run generates files for each command.
func (o *CmdOptions) Run(args []string) (err error) {
	defer func() { helper.RecordOSBuilderUsage("cmd", err) }()

	fm := file.NewFileManager(o.RootDir, o.Force)

	// Complete the CLI server metadata
	cli := o.Project.FindCLITool(o.BinaryName).Complete(o.Project)

	for _, kind := range o.Kinds {
		// Initialize metadata (using REST logic for naming conventions like Singular/Plural)
		cli.PrepareRESTMetadata(kind)

		// Generate files
		if err := o.GenerateFiles(fm, cli); err != nil {
			return err
		}

		// Optional: Auto-register imports if your project structure supports it (e.g., in a root.go or all.go)
		// This part depends heavily on your specific CLI framework structure.
		// Example:
		// importPath := fmt.Sprintf(`%s/internal/%s/cmd/%s`, cliServer.Proj.M.ModuleName, cliServer.Name, cliServer.Cmd.PackageName)
		// rootFile := filepath.Join(cliServer.CmdDir(), "root.go")
		// _ = fm.AddImportToFile(rootFile, "_", importPath)
		internalDir := filepath.Join(o.Project.M.WorkDir, fmt.Sprintf("internal/%s", cli.Name))
		allFile := fmt.Sprintf("%s/cmd/all/all.go", internalDir)

		importPath := fmt.Sprintf(`%s/internal/%s/cmd/%s`, cli.Proj.M.ModuleName, cli.Name, cli.R.SingularLower)
		if cli.R.ResourcePathPrefix != "" {
			importPath = fmt.Sprintf(`%s/internal/%s/cmd/%s/%s`, cli.Proj.M.ModuleName, cli.Name, cli.R.ResourcePathPrefix, cli.R.SingularLower)
		}
		if err := fm.AddImportToFile(allFile, "_", importPath); err != nil {
			return err
		}

	}

	if o.ShowTips {
		o.PrintGettingStarted(cli)
	}
	return nil
}

// GenerateFiles materializes files for the selected command.
func (o *CmdOptions) GenerateFiles(fm *file.FileManager, cli *types.CLITool) error {
	// Define mappings
	// We assume there is a generic CLI command template at /project/internal/cmd/cmd.go
	// and we want to place it in internal/<binary>/cmd/<command_name>/<command_name>.go
	// or internal/<binary>/cmd/<command_name>.go depending on your preference.

	// Example Layout: internal/dmsctl/cmd/fake/fake.go
	targetPath := filepath.Join(cli.CMDDir(), cli.R.SingularLower, cli.R.FileName)
	if cli.R.ResourcePathPrefix != "" {
		targetPath = filepath.Join(cli.CMDDir(), cli.R.ResourcePathPrefix, cli.R.SingularLower, cli.R.SingularLower+".go")
	}

	pairs := map[string]string{
		targetPath: "/project/internal/mbctl/cmd/template/template.go",
	}

	// Generate templated files
	if err := helper.RenderTemplate(
		fm,
		pairs,
		helper.GetTemplateFuncMap(),
		&types.TemplateData{Project: o.Project, CLI: cli},
	); err != nil {
		return err
	}

	return nil
}

// PrintGettingStarted prints follow-up instructions.
func (o *CmdOptions) PrintGettingStarted(cli *types.CLITool) {
	fmt.Printf("\n%s Command(s) creation succeeded %s\n", emoji.CheckMarkButton, color.GreenString("%s", strings.Join(o.Kinds, ",")))

	fmt.Printf("%s Next steps:\n", emoji.Parse(":rocket:"))

	if o.Project.Metadata.MakefileMode == known.MakefileModeNone {
		PrintClosingTips(o.Project.M.ProjectName)
		return
	}

	fmt.Printf("%s Use the following command to re-compile the project %s:\n\n", emoji.Parse(":computer:"), emoji.Parse(":point_down:"))

	fmt.Println(
		color.WhiteString("$ make build BINS=%s", cli.BinaryName),
		color.CyanString("# build %s", cli.BinaryName),
	)

	PrintClosingTips(o.Project.M.ProjectName)
}
