package util

import (
	"log/slog"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericiooptions"
)

// GroupName represents the category identifier for a set of commands.
// It is used to group commands visually in the help output.
type GroupName string

const (
	// GroupBasic contains fundamental commands for beginners.
	GroupBasic GroupName = "Basic Commands (Beginner):"

	// GroupProject contains commands related to project management.
	GroupProject GroupName = "Project Commands:"

	// GroupTroubleshooting contains commands for debugging and diagnostics.
	GroupTroubleshooting GroupName = "Troubleshooting and Debugging:"

	// GroupAdvanced contains commands for power users or low-level operations.
	GroupAdvanced GroupName = "Advanced Commands:"
)

// CommandConstructor defines the function signature for creating a cobra.Command.
// It injects dependencies like Factory and IOStreams.
// Note: Factory is assumed to be defined in this package or imported.
type CommandConstructor func(f Factory, ioStreams genericiooptions.IOStreams) *cobra.Command

type registryItem struct {
	group       GroupName
	constructor CommandConstructor
}

// registry stores all registered command constructors.
// It uses a slice to preserve the order of registration.
var registry = make([]registryItem, 0)

// Register adds a command constructor to the global registry under a specific group.
// This function is typically called in the init() function of subcommand packages.
func Register(group GroupName, constructor CommandConstructor) {
	// Log the registration event using structured logging.
	// Using "group" as a key instead of embedding it in the string.
	slog.Debug("registering command constructor", "group", group)

	registry = append(registry, registryItem{
		group:       group,
		constructor: constructor,
	})
}

// GetCommands instantiates and returns all registered commands, organized by their group.
// It filters out any nil commands returned by constructors.
func GetCommands(f Factory, ioStreams genericiooptions.IOStreams) map[GroupName][]*cobra.Command {
	result := make(map[GroupName][]*cobra.Command)

	for _, item := range registry {
		cmd := item.constructor(f, ioStreams)
		if cmd == nil {
			slog.Warn("command constructor returned nil", "group", item.group)
			continue
		}

		result[item.group] = append(result[item.group], cmd)
	}

	return result
}
