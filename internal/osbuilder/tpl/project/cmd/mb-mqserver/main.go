package main

import (
	"os"

	"{{.M.ModuleName}}/cmd/{{.MQ.BinaryName}}/app"
)

// The default entry point of a Go program. Serves as the starting point
// for reading the project code.
func main() {
	// Initialize the main command for the mqserver application.
	command := app.NewMQServerCommand()

	// Execute the command. If an error occurs, the program exits.
	// The exit code provides an indication of failure for external systems.
	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}
