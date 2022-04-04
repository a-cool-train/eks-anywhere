package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/aws/eks-anywhere/pkg/curatedpackages"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
)

type upgradePackageOptions struct {
	bundleVersion string
}

var upo = &upgradePackageOptions{}

func init() {
	upgradeCmd.AddCommand(upgradePackagesCommand)
	upgradePackagesCommand.Flags().StringVar(&upo.bundleVersion, "bundleversion", "", "Bundle version to use")
	err := upgradePackagesCommand.MarkFlagRequired("bundleversion")
	if err != nil {
		log.Fatalf("Error marking flag as required: %v", err)
	}
}

var upgradePackagesCommand = &cobra.Command{
	Use:          "packages",
	Short:        "Upgrade curated packages to the latest version",
	PreRunE:      preRunPackages,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := upgradePackages(cmd.Context()); err != nil {
			return err
		}
		return nil
	},
}

func upgradePackages(ctx context.Context) error {
	deps, err := newDependencies(ctx)
	if err != nil {
		return fmt.Errorf("unable to initialize executables: %v", err)
	}
	kubectl := deps.Kubectl
	kubeConfig := kubeconfig.FromEnvironment()
	activeController, err := curatedpackages.GetActiveController(ctx, kubectl, kubeConfig)
	if err != nil {
		return err
	}
	err = curatedpackages.UpgradeBundle(ctx, activeController, kubectl, upo.bundleVersion, kubeConfig)
	if err != nil {
		return err
	}
	return nil
}
