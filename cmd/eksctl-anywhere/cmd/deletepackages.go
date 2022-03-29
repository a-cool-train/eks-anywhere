package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/aws/eks-anywhere/pkg/curatedpackages"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
)

func init() {
	deleteCmd.AddCommand(deletePackageCommand)
}

var deletePackageCommand = &cobra.Command{
	Use:          "package(s) [flags]",
	Aliases:      []string{"package", "packages"},
	Short:        "Delete package(s)",
	Long:         "This command is used to delete the curated packages installed in the cluster",
	PreRunE:      preRunPackages,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return deleteResources(cmd.Context(), args)
	},
}

func deleteResources(ctx context.Context, args []string) error {
	kubeConfig := kubeconfig.FromEnvironment()
	err := curatedpackages.DeletePackages(ctx, args, kubeConfig)
	return err
}
