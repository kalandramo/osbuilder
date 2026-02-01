{{- $D := or .Web .MQ .Job -}}
package main

import (
    "os"
 
    "{{.M.ModuleName}}/cmd/{{$D.BinaryName}}/app"  
)

// The default entry point of a Go program. Serves as the starting point
// for reading the project code.
func main() {
    {{- if .Web }}
    // Initialize the main command for the apiserver application.
    command := app.NewAPIServerCommand()
    {{- else if .Job }}
    // Initialize the main command for the jobserver application.
    command := app.NewJobServerCommand()
    {{- end }}
 
    // Execute the command. If an error occurs, the program exits.
    // The exit code provides an indication of failure for external systems.
    if err := command.Execute(); err != nil {
        os.Exit(1)
    }
}
