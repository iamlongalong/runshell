package main

import (
	"github.com/iamlongalong/runshell/cmd/runshell/cmd"
	_ "github.com/iamlongalong/runshell/cmd/runshell/docs" // Import generated Swagger docs
)

func main() {
	cmd.Execute()
}
