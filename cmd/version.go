package cmd

import (
	ver "github.com/ccustine/beastie/version"
	"github.com/spf13/cobra"
)

func NewVersionCmd() *cobra.Command {
	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the version number",
		Long:  `Shows the version number if a final release, or a commit and date for snapshots`,
		Run: func(cmd *cobra.Command, args []string) {
			//fmt.Println("The version is: ", VERSION)
			ver.PrintVersion()
		},
	}

	return versionCmd
}
