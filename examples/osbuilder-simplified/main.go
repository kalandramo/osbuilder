package main

import (
	"os"

	"github.com/onexstack/osbuilder/examples/osbuilder-simplified/internal/cmd"
)

func main() {
	command := cmd.NewDefaultOSCtlCommand()
	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}
