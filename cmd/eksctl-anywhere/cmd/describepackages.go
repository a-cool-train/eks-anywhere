package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/aws/eks-anywhere/pkg/curatedpackages"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
)

func init() {
	describeCmd.AddCommand(describePackagesCommand)
}

var describePackagesCommand = &cobra.Command{
	Use:          "packages",
	Short:        "Describe curated packages",
	PreRunE:      preRunPackages,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := describeResources(cmd.Context(), args); err != nil {
			return err
		}
		return nil
	},
}

func describeResources(ctx context.Context, args []string) error {
	kubeConfig := kubeconfig.FromEnvironment()
	err := curatedpackages.DescribePackages(ctx, args, kubeConfig)
	return err
}
