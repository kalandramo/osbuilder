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

// MQOptions holds flags and runtime context for the 'create mq' command.
type MQOptions struct {
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
	mqLongDesc = templates.LongDesc(`
        Create Message Queue (MQ) consumers and handlers for your project.

        This command scaffolds MQ artifacts (proto, handlers, biz logic, store) for the given message kinds/topics.`)

	mqExamples = templates.Examples(`
        # Create MQ consumer for a specific topic/kind
        osbuilder create mq --kinds job --binary-name mb-mqserver

        # Create multiple consumers
        osbuilder create mq --kinds cron_job,job --binary-name mb-mqserver`)
)

// NewMQOptions creates a default MQOptions.
func NewMQOptions(io genericiooptions.IOStreams) *MQOptions {
	return &MQOptions{
		APIVersion: "v1",
		ShowTips:   true,
		IOStreams:  io,
	}
}

// NewCmdMQ builds the 'create mq' cobra command.
func NewCmdMQ(factory cmdutil.Factory, ioStreams genericiooptions.IOStreams) *cobra.Command {
	o := NewMQOptions(ioStreams)

	cmd := &cobra.Command{
		Use:                   "mq",
		DisableFlagsInUseLine: true,
		Short:                 "Create MQ consumers",
		Long:                  mqLongDesc,
		Example:               mqExamples,
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
	cmd.Flags().StringVarP(&o.BinaryName, "binary-name", "b", o.BinaryName, "Target binary/worker name (e.g., mb-mqserver).")
	cmd.Flags().BoolVarP(&o.Force, "force", "f", o.Force, "Force overwriting of existing files.")
	// Add hidden flags
	cmd.Flags().StringVar(&o.RootDir, "root-dir", "", "Override root directory (hidden flag)")
	cmd.Flags().BoolVar(&o.ShowTips, "show-tips", o.ShowTips, "Print post-run tips.")
	_ = cmd.Flags().MarkHidden("root-dir")
	_ = cmd.Flags().MarkHidden("show-tips")

	return cmd
}

// Complete resolves working directory and loads project metadata.
func (o *MQOptions) Complete(factory cmdutil.Factory, cmd *cobra.Command, args []string) error {
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
	if o.BinaryName == "" && len(proj.MQServers) == 1 {
		o.BinaryName = proj.MQServers[0].Name
	}

	o.Project = proj
	return nil
}

// Validate checks required inputs and project state.
func (o *MQOptions) Validate(cmd *cobra.Command, args []string) error {
	if o.Project == nil {
		return fmt.Errorf("project not loaded")
	}
	if len(o.Kinds) == 0 {
		return fmt.Errorf("at least one kind/topic must be provided via --kinds")
	}
	mq := o.Project.FindMQServer(o.BinaryName)
	if mq == nil {
		return fmt.Errorf("mq server/binary %q not found in project; use --binary-name", o.BinaryName)
	}
	return nil
}

// Run generates files for each kind and updates related components.
func (o *MQOptions) Run(args []string) (err error) {
	defer func() { helper.RecordOSBuilderUsage("mq", err) }()

	fm := file.NewFileManager(o.RootDir, o.Force)

	mq := o.Project.FindMQServer(o.BinaryName).Complete(o.Project)
	for _, kind := range o.Kinds {
		// Initialize metadata (using REST logic for naming conventions like Singular/Plural)
		mq.PrepareRESTMetadata(kind)

		// Generate files (consumer, handler, store, biz, model)
		if err := o.GenerateFiles(fm, mq); err != nil {
			return err
		}

		// Update store.go (Assuming the MQ consumer needs database access)
		internalDir := filepath.Join(o.Project.M.WorkDir, fmt.Sprintf("internal/%s", mq.Name))
		if err := fm.AddNewMethod("store", filepath.Join(internalDir, "store", "store.go"), mq, ""); err != nil {
			return err
		}

		importPathSuffix := fmt.Sprintf("%s", mq.R.SingularLower)
		if mq.R.ResourcePathPrefix != "" {
			importPathSuffix = fmt.Sprintf("%s/%s", mq.R.ResourcePathPrefix, mq.R.SingularLower)
		}
		// Update biz.go (Assuming the MQ consumer calls business logic)
		if err := fm.AddNewMethod(
			"biz",
			filepath.Join(internalDir, "biz", "biz.go"),
			mq,
			fmt.Sprintf("%s/internal/%s/biz/%s/%s",
				o.Project.M.ModuleName,
				mq.Name,
				o.Project.M.APIVersion,
				importPathSuffix,
			),
		); err != nil {
			return err
		}
	}

	if o.ShowTips {
		o.PrintGettingStarted(mq)
	}
	return nil
}

// GenerateFiles materializes files for the selected web server and kind.
func (o *MQOptions) GenerateFiles(fm *file.FileManager, mq *types.MQServer) error {
	// Define the mapping between local file paths and the template source paths.
	// Note: The template source paths (values) point to specific MQ templates in the asset system.
	pairs := map[string]string{
		filepath.Join(mq.APIDir(), mq.R.SingularLower+".proto"):      "/project/pkg/api/mqserver/v1/post.proto",
		filepath.Join(mq.RESTBizFile()):                              "/project/internal/mqserver/biz/v1/post/post.go",
		filepath.Join(mq.StoreDir(), mq.R.FileName):                  "/project/internal/mqserver/store/post.go",
		filepath.Join(mq.PkgDir(), "validation", mq.R.FileName):      "/project/internal/mqserver/pkg/validation/post.go",
		filepath.Join(mq.PkgDir(), "conversion", mq.R.FileName):      "/project/internal/mqserver/pkg/conversion/post.go",
		filepath.Join(mq.ModelDir(), mq.R.FileName):                  "/project/internal/mqserver/model/post.gen.go",
		filepath.Join(mq.ModelDir(), "hook_"+mq.R.FileName):          "/project/internal/mqserver/model/hook_post.go",
		filepath.Join(mq.Proj.InternalPkg(), "errno", mq.R.FileName): "/project/internal/pkg/errno/post.go",
	}

	switch mq.MQFramework {
	case known.MQFrameworkKafka:
		pairs[filepath.Join(mq.HandlerDir(), mq.R.FileName)] = "/project/internal/mqserver/handler/kafka/post.go"
	default:
	}

	// Generate templated files using the provided template engine
	if err := helper.RenderTemplate(
		fm,
		pairs,
		helper.GetTemplateFuncMap(),
		&types.TemplateData{Project: o.Project, MQ: mq},
	); err != nil {
		return err
	}

	return nil
}

// PrintGettingStarted prints follow-up commands to rebuild and run.
func (o *MQOptions) PrintGettingStarted(mq *types.MQServer) {
	fmt.Printf("\n%s MQ Consumer(s) creation succeeded %s\n", emoji.CheckMarkButton, color.GreenString("%s", strings.Join(o.Kinds, ",")))
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
		color.WhiteString("$ make build BINS=%s", mq.BinaryName),
		color.CyanString("# build %s", mq.BinaryName),
	)
	fmt.Println(color.WhiteString("After restarting, check logs to verify the MQ consumer is subscribed to the topic."))

	PrintClosingTips(o.Project.M.ProjectName)
}
