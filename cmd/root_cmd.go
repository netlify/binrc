package cmd

import (
	"fmt"
	"os"

	"github.com/netlify/binrc/cache"
	"github.com/spf13/cobra"
)

// rootCmd is the main Binrc command.
// It runs `use` when no subcommand specified.
var rootCmd = &cobra.Command{
	Use:  "binrc",
	Long: "A command line application to manage different versions of binaries on GitHub releases",
	Run:  execInstallCmd,
}

// RootCmd adds flags and subcommands to the root command.
func RootCmd() *cobra.Command {
	rootCmd.PersistentFlags().StringP("-cache-store-path", "c", cache.DefaultStorePath, "The path to binrc's cache directory, $HOME/.binrc by default")
	rootCmd.AddCommand(installCmd, versionCmd)
	return rootCmd
}

func displayError(err error) {
	if os.Getenv("DEBUG") != "" {
		fmt.Printf("%+v\n", err)
	} else {
		fmt.Println(err)
	}

	os.Exit(1)
}
