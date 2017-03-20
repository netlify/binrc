package cmd

import (
	"fmt"
	"path"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/netlify/binrc/cache"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:  "install [PROJECT_PATH] [VERSION]",
	Long: "This subcommand installs the binary in your system.",
	Run:  execInstallCmd,
}

func execInstallCmd(cmd *cobra.Command, args []string) {
	sp := cmd.Flag("cache-store-path").Value.String()
	h, err := homedir.Dir()
	if err != nil {
		displayError(err)
		return
	}

	c := cache.New(path.Clean(path.Join(sp, h)))
	pc, err := c.GetOrSet(args[0], args[1])
	if err != nil {
		displayError(err)
	}

	fmt.Println(pc.BinaryPath)
}
