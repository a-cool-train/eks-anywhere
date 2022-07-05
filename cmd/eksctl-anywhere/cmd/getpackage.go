package cmd

import (
	"github.com/spf13/cobra"
)

type getPackageOptions struct {
	output     string
	useLibrary bool
}

var gpo = &getPackageOptions{}

func init() {
	getCmd.AddCommand(getPackageCommand)
	getPackageCommand.Flags().StringVarP(&gpo.output, "output", "o", "", "Specifies the output format (valid option: json, yaml)")
	getPackageCommand.Flags().BoolVarP(&gpo.useLibrary, "use-library", "u", false, "Specifies whether to use library or container")
}

var getPackageCommand = &cobra.Command{
	Use:          "package(s) [flags]",
	Aliases:      []string{"package", "packages"},
	Short:        "Get package(s)",
	Long:         "This command is used to display the curated packages installed in the cluster",
	PreRunE:      preRunPackages,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return getResources(cmd.Context(), "packages", gpo.output, args, gpo.useLibrary)
	},
}
