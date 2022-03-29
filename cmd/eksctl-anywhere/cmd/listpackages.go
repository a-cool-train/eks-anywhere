package cmd

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/aws/eks-anywhere/pkg/curatedpackages"
)

type listPackagesOption struct {
	kubeVersion  curatedpackages.KubeVersion
	bundleSource curatedpackages.BundleSource
}

var lpo = &listPackagesOption{}

func init() {
	listCmd.AddCommand(listPackagesCommand)
	listPackagesCommand.Flags().Var(&lpo.bundleSource, "source", "Discovery Location. Options (cluster, registry)")
	listPackagesCommand.Flags().Var(&lpo.kubeVersion, "kubeversion", "Kubernetes Version of the cluster to be used. Format <major>.<minor>")
	err := listPackagesCommand.MarkFlagRequired("source")
	if err != nil {
		log.Fatalf("Error marking flag as required: %v", err)
	}
}

var listPackagesCommand = &cobra.Command{
	Use:          "packages",
	Short:        "Lists curated packages available to install",
	PreRunE:      preRunPackages,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := curatedpackages.ListPackages(cmd.Context(), lpo.bundleSource, lpo.kubeVersion); err != nil {
			return err
		}
		return nil
	},
}
