package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version is the Git SHA specified when Binrc was build.
var Version string

var versionCmd = &cobra.Command{
	Use:  "version",
	Long: "This subcommand displays Binrc's current version.",
	Run:  showVersion,
}

func showVersion(cmd *cobra.Command, args []string) {
	if Version == "" {
		fmt.Println("Unknown version, this binary was probably built by `go get`")
	} else {
		fmt.Println(Version)
	}
}
