package cmd

import "github.com/spf13/cobra"

var useCmd = &cobra.Command{
	Use:  "use",
	Long: "This subcommand sets the right binary in your path to be executed later.",
	Run:  execUseCmd,
}

func execUseCmd(cmd *cobra.Command, args []string) {
}
