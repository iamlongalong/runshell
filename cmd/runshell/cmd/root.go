package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	auditDir    string
	httpAddr    string
	dockerImage string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "runshell",
	Short: "A powerful command executor",
	Long: `RunShell is a powerful command executor that supports local and Docker execution,
with built-in commands, audit logging, and HTTP server capabilities.

Example:
  runshell exec ls -l
  runshell exec --docker-image alpine:latest ls -l
  runshell server --http :8080`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&auditDir, "audit-dir", "", "Directory for audit logs")
	rootCmd.PersistentFlags().StringVar(&dockerImage, "docker-image", "ubuntu:latest", "Docker image to use for container execution")
}
