package cmd

import (
	"context"
	"log"

	"github.com/spf13/cobra"

	"github.com/aws/eks-anywhere/pkg/curatedpackages"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
)

type installPackageOptions struct {
	source      curatedpackages.BundleSource
	kubeVersion string
	name        string
}

var ipo = &installPackageOptions{}

func init() {
	installCmd.AddCommand(installPackageCommand)
	installPackageCommand.Flags().Var(&ipo.source, "source", "Location to find curated packages: (cluster, registry)")
	installPackageCommand.Flags().StringVar(&ipo.kubeVersion, "kubeversion", "", "Kubernetes Version of the cluster to be used. Format <major>.<minor>")
	installPackageCommand.Flags().StringVar(&ipo.name, "name", "", "Custom name of the curated package to install")
	if err := installPackageCommand.MarkFlagRequired("source"); err != nil {
		log.Fatalf("Error marking flag as required: %v", err)
	}
	if err := installPackageCommand.MarkFlagRequired("name"); err != nil {
		log.Fatalf("Error marking flag as required: %v", err)
	}
}

var installPackageCommand = &cobra.Command{
	Use:          "packages [flags]",
	Aliases:      []string{"package", "packages"},
	Short:        "Install package(s)",
	Long:         "This command is used to Install a curated package. Use list to discover curated packages",
	PreRunE:      preRunPackages,
	SilenceUsage: true,
	RunE:         runInstallPackages(),
}

func runInstallPackages() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := validateKubeVersion(ipo.kubeVersion, ipo.source); err != nil {
			return err
		}

		return installPackages(cmd.Context(), ipo, args)
	}
}

func installPackages(ctx context.Context, ipo *installPackageOptions, args []string) error {
	kubeConfig := kubeconfig.FromEnvironment()
	bundle, err := curatedpackages.GetLatestBundle(ctx, kubeConfig, ipo.source, ipo.kubeVersion)
	if err != nil {
		return err
	}
	p, err := curatedpackages.GetPackageFromBundle(bundle, args[0])
	if err != nil {
		return err
	}
	err = curatedpackages.InstallPackage(ctx, p, bundle, ipo.name, kubeConfig)
	if err != nil {
		return err
	}
	return nil
}
