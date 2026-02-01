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

// JobOptions holds flags and runtime context for the 'create job' command.
type JobOptions struct {
	RootDir string

	Kinds      []string // Message kinds/topics to generate (snake_case recommended)
	BinaryName string   // Target web server/binary name
	Force      bool     // Overwrite files if they exist

	// APIVersion is kept for directory structure consistency (e.g. internal/app/biz/v1)
	APIVersion string
	ShowTips   bool // Print getting-started hints

	Project *types.Project // Loaded project metadata

	genericiooptions.IOStreams
}

var (
	jobLongDesc = templates.LongDesc(`
        Create Message Queue (Job) consumers and handlers for your project.

        This command scaffolds Job artifacts (proto, handlers, biz logic, store) for the given message kinds/topics.`)

	jobExamples = templates.Examples(`
        # Create Job consumer for a specific topic/kind
        osbuilder create job --kinds job --binary-name mb-jobserver

        # Create multiple consumers
        osbuilder create job --kinds cron_job,job --binary-name mb-jobserver`)
)

// NewJobOptions creates a default JobOptions.
func NewJobOptions(io genericiooptions.IOStreams) *JobOptions {
	return &JobOptions{
		APIVersion: "v1",
		ShowTips:   true,
		IOStreams:  io,
	}
}

// NewCmdJob builds the 'create job' cobra command.
func NewCmdJob(factory cmdutil.Factory, ioStreams genericiooptions.IOStreams) *cobra.Command {
	o := NewJobOptions(ioStreams)

	cmd := &cobra.Command{
		Use:                   "job",
		DisableFlagsInUseLine: true,
		Short:                 "Create a new asynchronous job",
		Long:                  jobLongDesc,
		Example:               jobExamples,
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
	cmd.Flags().StringSliceVarP(&o.Kinds, "kinds", "", o.Kinds, "Message kinds/topics to generate in snake_case (e.g., post).")
	cmd.Flags().StringVarP(&o.BinaryName, "binary-name", "b", o.BinaryName, "Target binary/worker name (e.g., mb-jobserver).")
	cmd.Flags().BoolVarP(&o.Force, "force", "f", o.Force, "Force overwriting of existing files.")
	// Add hidden flags
	cmd.Flags().StringVar(&o.RootDir, "root-dir", "", "Override root directory (hidden flag)")
	cmd.Flags().BoolVar(&o.ShowTips, "show-tips", o.ShowTips, "Print post-run tips.")
	_ = cmd.Flags().MarkHidden("root-dir")
	_ = cmd.Flags().MarkHidden("show-tips")

	return cmd
}

// Complete resolves working directory and loads project metadata.
func (o *JobOptions) Complete(factory cmdutil.Factory, cmd *cobra.Command, args []string) error {
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
		APIVersion: "v1", // Defaulting to v1 for directory structures
		APIAlias:   "v1",
		ModuleName: MustModulePath(proj.Metadata.ModulePath, o.RootDir),
	}).Complete()
	proj.M.ProjectName = filepath.Base(o.RootDir)
	proj.M.RegistryPrefix = proj.Metadata.Image.RegistryPrefix

	// If a single web server exists and BinaryName not set, default to it.
	if o.BinaryName == "" && len(proj.JobServers) == 1 {
		o.BinaryName = proj.JobServers[0].Name
	}

	o.Project = proj
	return nil
}

// Validate checks required inputs and project state.
func (o *JobOptions) Validate(cmd *cobra.Command, args []string) error {
	if o.Project == nil {
		return fmt.Errorf("project not loaded")
	}
	if len(o.Kinds) == 0 {
		return fmt.Errorf("at least one kind/topic must be provided via --kinds")
	}
	job := o.Project.FindJobServer(o.BinaryName)
	if job == nil {
		return fmt.Errorf("job server/binary %q not found in project; use --binary-name", o.BinaryName)
	}
	return nil
}

// Run generates files for each kind and updates related components.
func (o *JobOptions) Run(args []string) (err error) {
	defer func() { helper.RecordOSBuilderUsage("job", err) }()

	fm := file.NewFileManager(o.RootDir, o.Force)

	job := o.Project.FindJobServer(o.BinaryName).Complete(o.Project)
	for _, kind := range o.Kinds {
		// Initialize metadata (using REST logic for naming conventions like Singular/Plural)
		job.PrepareRESTMetadata(kind)

		// Generate files (consumer, handler, store, biz, model)
		if err := o.GenerateFiles(fm, job); err != nil {
			return err
		}

		// Update store.go (Assuming the Job consumer needs database access)
		internalDir := filepath.Join(o.Project.M.WorkDir, fmt.Sprintf("internal/%s", job.Name))
		if err := fm.AddNewMethod("store", filepath.Join(internalDir, "store", "store.go"), job, ""); err != nil {
			return err
		}

		allFile := fmt.Sprintf("%s/watcher/all/all.go", internalDir)
		importPath := fmt.Sprintf(`%s/internal/%s/watcher/customized/%s`, job.Proj.M.ModuleName, job.Name, job.R.SingularLower)
		if err := fm.AddImportToFile(allFile, "//_", importPath); err != nil {
			return err
		}
	}

	if o.ShowTips {
		o.PrintGettingStarted(job)
	}
	return nil
}

// GenerateFiles materializes files for the selected web server and kind.
func (o *JobOptions) GenerateFiles(fm *file.FileManager, job *types.JobServer) error {
	// Define the mapping between local file paths and the template source paths.
	// Note: The template source paths (values) point to specific Job templates in the asset system.
	pairs := map[string]string{
		filepath.Join(job.StoreDir(), job.R.FileName):                                                 "/project/internal/jobserver/store/post.go",
		filepath.Join(job.ModelDir(), job.R.FileName):                                                 "/project/internal/jobserver/model/post.gen.go",
		filepath.Join(job.ModelDir(), "hook_"+job.R.FileName):                                         "/project/internal/jobserver/model/hook_post.go",
		filepath.Join(job.WatcherDir(), fmt.Sprintf("customized/%s/watcher.go", job.R.SingularLower)): "/project/internal/jobserver/watcher/customized/post/watcher.go",
		// filepath.Join(job.Proj.InternalPkg(), "errno", job.R.FileName): "/project/internal/pkg/errno/post.go",
	}

	// Generate templated files using the provided template engine
	if err := helper.RenderTemplate(
		fm,
		pairs,
		helper.GetTemplateFuncMap(),
		&types.TemplateData{Project: o.Project, Job: job},
	); err != nil {
		return err
	}

	return nil
}

// PrintGettingStarted prints follow-up commands to rebuild and run.
func (o *JobOptions) PrintGettingStarted(job *types.JobServer) {
	fmt.Printf("\n%s Job Consumer(s) creation succeeded %s\n", emoji.CheckMarkButton, color.GreenString("%s", strings.Join(o.Kinds, ",")))
	if o.Project.Metadata.MakefileMode == known.MakefileModeNone {
		PrintClosingTips(o.Project.M.ProjectName)
		return
	}

	fmt.Printf("%s Use the following command to re-compile the project %s:\n\n", emoji.Parse(":computer:"), emoji.Parse(":point_down:"))

	fmt.Println(
		color.WhiteString("$ cd %s", o.RootDir),
		color.CyanString("# enter project directory"),
	)
	fmt.Println(
		color.WhiteString("$ make build BINS=%s", job.BinaryName),
		color.CyanString("# build %s", job.BinaryName),
	)
	fmt.Println(color.WhiteString("After restarting, check logs to verify the Job consumer is subscribed to the topic."))

	PrintClosingTips(o.Project.M.ProjectName)
}
