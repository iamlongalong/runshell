package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "runshell",
	Short: "A modern shell command executor",
	Long: `RunShell is a modern shell command executor that supports:
- Local and Docker command execution
- HTTP API for remote execution
- Command piping and scripting
- Audit logging and security controls`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
