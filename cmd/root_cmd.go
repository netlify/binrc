package cmd

import "github.com/spf13/cobra"

// rootCmd is the main Binrc command.
// It runs `use` when no subcommand specified.
var rootCmd = &cobra.Command{
	Use:  "binrc",
	Long: "A command line application to manage different versions of binaries on GitHub releases",
	Run:  execUseCmd,
}

// RootCmd adds flags and subcommands to the root command.
func RootCmd() *cobra.Command {
	rootCmd.PersistentFlags().StringP("cache-path", "c", "", "The path to binrc's cache directory, $HOME/.binrc by default")
	rootCmd.AddCommand(useCmd, versionCmd)
	return rootCmd
}
