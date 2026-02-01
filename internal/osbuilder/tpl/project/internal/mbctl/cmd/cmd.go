package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/onexstack/onexstack/pkg/cli/cli"
	"github.com/onexstack/onexstack/pkg/core"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/client-go/rest"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog/v2"
	genericcmd "k8s.io/kubectl/pkg/cmd"
	"k8s.io/kubectl/pkg/cmd/plugin"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
	"k8s.io/kubectl/pkg/util/term"

	_ "{{.M.ModuleName}}/internal/{{.CLI.Name}}/cmd/all"
	"{{.M.ModuleName}}/internal/{{.CLI.Name}}/cmd/options"
	cmdutil "{{.M.ModuleName}}/internal/{{.CLI.Name}}/cmd/util"
	"{{.M.ModuleName}}/internal/{{.CLI.Name}}/cmd/version"
	clioptions "{{.M.ModuleName}}/internal/{{.CLI.Name}}/util/options"
)

const (
	{{.CLI.Name}}CmdHeaders  = "{{.CLI.Name | toupper}}_COMMAND_HEADERS"

    // defaultHomeDir defines the default directory to store configuration files
    // for the {{.CLI.BinaryName}} service, typically within the user's home directory.
    defaultHomeDir = ".{{.M.ProjectName}}"
 
    // defaultConfigName specifies the default configuration file name
    // for the {{.CLI.BinaryName}} service.
    defaultConfigName = "{{.CLI.BinaryName}}.yaml"
)


type CLIToolOptions struct {
	PluginHandler genericcmd.PluginHandler
	Arguments     []string

	genericiooptions.IOStreams
}

// configFile stores the path to the configuration file, set via command-line flag.
var configFile string

// NewDefaultCLIToolCommand creates the `{{.CLI.Name}}` command with default arguments.
func NewDefaultCLIToolCommand() *cobra.Command {
	ioStreams := genericiooptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}
	return NewDefaultCLIToolCommandWithArgs(CLIToolOptions{
		PluginHandler: genericcmd.NewDefaultPluginHandler([]string{"{{.CLI.Name}}"}),
		Arguments:     os.Args,
		IOStreams:     ioStreams,
	})
}

// NewDefaultCLIToolCommandWithArgs creates the `{{.CLI.Name}}` command with arguments.
func NewDefaultCLIToolCommandWithArgs(o CLIToolOptions) *cobra.Command {
	cmd := NewCLIToolCommand(o)

	if o.PluginHandler == nil {
		return cmd
	}

	if len(o.Arguments) > 1 {
		cmdPathPieces := o.Arguments[1:]

		// only look for suitable extension executables if
		// the specified command does not already exist
		if foundCmd, foundArgs, err := cmd.Find(cmdPathPieces); err != nil {
			// Also check the commands that will be added by Cobra.
			// These commands are only added once rootCmd.Execute() is called, so we
			// need to check them explicitly here.
			var cmdName string // first "non-flag" arguments
			for _, arg := range cmdPathPieces {
				if !strings.HasPrefix(arg, "-") {
					cmdName = arg
					break
				}
			}

			switch cmdName {
			case "help", cobra.ShellCompRequestCmd, cobra.ShellCompNoDescRequestCmd:
				// Don't search for a plugin
			default:
				if err := genericcmd.HandlePluginCommand(o.PluginHandler, cmdPathPieces, 1); err != nil {
					fmt.Fprintf(o.IOStreams.ErrOut, "Error: %v\n", err)
					os.Exit(1)
				}
			}
		} else if err == nil {
			if !cmdutil.CmdPluginAsSubcommand.IsDisabled() {
				// Command exists(e.g. kubectl create), but it is not certain that
				// subcommand also exists (e.g. kubectl create networkpolicy)
				// we also have to eliminate kubectl create -f
				if genericcmd.IsSubcommandPluginAllowed(foundCmd.Name()) && len(foundArgs) >= 1 && !strings.HasPrefix(foundArgs[0], "-") {
					subcommand := foundArgs[0]
					builtinSubcmdExist := false
					for _, subcmd := range foundCmd.Commands() {
						if subcmd.Name() == subcommand {
							builtinSubcmdExist = true
							break
						}
					}

					if !builtinSubcmdExist {
						if err := genericcmd.HandlePluginCommand(o.PluginHandler, cmdPathPieces, len(cmdPathPieces)-len(foundArgs)+1); err != nil {
							fmt.Fprintf(o.IOStreams.ErrOut, "Error: %v\n", err)
							os.Exit(1)
						}
					}
				}
			}
		}
	}

	return cmd
}

// NewCLIToolCommand creates the `{{.CLI.Name}}` command and its nested children.
func NewCLIToolCommand(o CLIToolOptions) *cobra.Command {
	warningHandler := rest.NewWarningWriter(o.IOStreams.ErrOut, rest.WarningWriterOptions{Deduplicate: true, Color: term.AllowsColorOutput(o.IOStreams.ErrOut)})
	warningsAsErrors := false
	opts := clioptions.NewCLIToolOptions()
	// Parent command to which all subcommands are added.
	cmds := &cobra.Command{
		Use:   "{{.CLI.Name}}",
		Short: i18n.T("{{.CLI.Name}} is a command-line tool for the onex technology stack scaffold"),
		Long: templates.LongDesc(`
        {{.CLI.Name}} is a command-line tool for the onex technology stack scaffold..`),
		Run: runHelp,
		// Hook before and after Run initialize and write profiles to disk,
		// respectively.
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			rest.SetDefaultWarningHandler(warningHandler)

			if cmd.Name() == cobra.ShellCompRequestCmd {
				// This is the __complete or __completeNoDesc command which
				// indicates shell completion has been requested.
				plugin.SetupPluginCompletion(cmd, args)
			}

			// Unmarshal the configuration from viper into opts
            if err := viper.Unmarshal(opts); err != nil {
                return fmt.Errorf("failed to unmarshal configuration: %w", err)
            }
         
            // Complete the options by setting default values and derived configurations
            if err := opts.Complete(); err != nil {
                return fmt.Errorf("failed to complete options: %w", err)
            }
         
            // Validate command-line options
            if err := opts.Validate(); err != nil {
                return fmt.Errorf("invalid options: %w", err)
            }

			return initProfiling()
		},
		PersistentPostRunE: func(*cobra.Command, []string) error {
			if err := flushProfiling(); err != nil {
				return err
			}
			if warningsAsErrors {
				count := warningHandler.WarningCount()
				switch count {
				case 0:
					// no warnings
				case 1:
					return fmt.Errorf("%d warning received", count)
				default:
					return fmt.Errorf("%d warnings received", count)
				}
			}
			return nil
		},
	}
	// From this point and forward we get warnings on flags that contain "_" separators
	// when adding them with hyphen instead of the original name.
	cmds.SetGlobalNormalizationFunc(cliflag.WarnWordSepNormalizeFunc)

	flags := cmds.PersistentFlags()

	addProfilingFlags(flags)

	flags.BoolVar(&warningsAsErrors, "warnings-as-errors", warningsAsErrors, "Treat warnings received from the server as errors and exit with a non-zero exit code")

	// Updates hooks to add onexctl command headers: SIG CLI KEP 859.
	addCmdHeaderHooks(cmds, opts)

	// Normalize all flags that are coming from other packages or pre-configurations
	// a.k.a. change all "_" to "-". e.g. glog package
	flags.SetNormalizeFunc(cliflag.WordSepNormalizeFunc)

    // Register the configuration initialization function, which runs before command execution.
    // It sets up Viper to search for configuration files in specified directories.
    cobra.OnInitialize(core.OnInitialize(&configFile, "{{.CLI.EnvironmentPrefix}}", cli.SearchDirs(defaultHomeDir), defaultConfigName))

    // Define persistent flags that apply to this command and its subcommands.
    flags.StringVarP(
        &configFile,
        "config",
        "c",
        cli.FilePath(defaultHomeDir, defaultConfigName),
        "Path to the {{.CLI.BinaryName}} configuration file.",
    )
 
    // Add server-specific options as command-line flags.
    opts.AddFlags(flags)

	f, _ := cmdutil.NewFactory(opts)

	registeredCmds := cmdutil.GetCommands(f, o.IOStreams)

	groups := templates.CommandGroups{
        {
            Message:  string(cmdutil.GroupBasic),
            Commands: []*cobra.Command{},
        },
        {
            Message:  string(cmdutil.GroupProject),
            Commands: []*cobra.Command{},
        },
        {
            Message:  string(cmdutil.GroupTroubleshooting),
            Commands: []*cobra.Command{},
        },
        {
            Message:  string(cmdutil.GroupAdvanced),
            Commands: []*cobra.Command{},
        },
	}

	for i := range groups {
		groupName := cmdutil.GroupName(groups[i].Message)
		if cmds, ok := registeredCmds[groupName]; ok {
			groups[i].Commands = append(groups[i].Commands, cmds...)
		}
	}

	groups.Add(cmds)

	filters := []string{"options"}

	// Hide the "alpha" subcommand if there are no alpha commands in this build.
	alpha := NewCmdAlpha(f, o.IOStreams)
	if !alpha.HasSubCommands() {
		filters = append(filters, alpha.Name())
	}

	// Add plugin command group to the list of command groups.
	// The commands are only injected for the scope of showing help and completion, they are not
	// invoked directly.
	pluginCommandGroup := plugin.GetPluginCommandGroup(cmds)
	groups = append(groups, pluginCommandGroup)

	templates.ActsAsRootCommand(cmds, filters, groups...)

	cmds.AddCommand(alpha)
	// cmds.AddCommand(cmdconfig.NewCmdConfig(f, clientcmd.NewDefaultPathOptions(), o.IOStreams))
	cmds.AddCommand(plugin.NewCmdPlugin(o.IOStreams))
	cmds.AddCommand(version.NewCmdVersion(f, o.IOStreams))
	cmds.AddCommand(options.NewCmdOptions(o.IOStreams.Out))

	// Stop warning about normalization of flags. That makes it possible to
	// add the klog flags later.
	cmds.SetGlobalNormalizationFunc(cliflag.WordSepNormalizeFunc)
	return cmds
}

// addCmdHeaderHooks performs updates on two hooks:
//  1. Modifies the passed "cmds" persistent pre-run function to parse command headers.
//     These headers will be subsequently added as X-headers to every
//     REST call.
//  2. Adds CommandHeaderRoundTripper as a wrapper around the standard
//     RoundTripper. CommandHeaderRoundTripper adds X-Headers then delegates
//     to standard RoundTripper.
//
// For beta, these hooks are updated unless the ONEX_COMMAND_HEADERS environment variable
// is set, and the value of the env var is false (or zero).
// See SIG CLI KEP 859 for more information:
//
//	https://github.com/kubernetes/enhancements/tree/master/keps/sig-cli/859-kubectl-headers
func addCmdHeaderHooks(cmds *cobra.Command, _ *clioptions.CLIToolOptions) {
	// If the feature gate env var is set to "false", then do no add kubectl command headers.
	if value, exists := os.LookupEnv({{.CLI.Name}}CmdHeaders); exists {
		if value == "false" || value == "0" {
			klog.V(5).Infoln("onexctl command headers turned off")
			return
		}
	}
	klog.V(5).Infoln("onexctl command headers turned on")
	crt := &genericclioptions.CommandHeaderRoundTripper{}
	existingPreRunE := cmds.PersistentPreRunE
	// Add command parsing to the existing persistent pre-run function.
	cmds.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		crt.ParseCommandHeaders(cmd, args)
		return existingPreRunE(cmd, args)
	}
}

func runHelp(cmd *cobra.Command, args []string) {
	_ = cmd.Help()
}

// searchDirs 返回默认的配置文件搜索目录.
func searchDirs() []string {
	// 获取用户主目录
	homeDir, err := os.UserHomeDir()
	// 如果获取用户主目录失败，则打印错误信息并退出程序
	cobra.CheckErr(err)
	return []string{filepath.Join(homeDir, defaultHomeDir), "."}
}
