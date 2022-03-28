package cmd

import (
	"context"
	"github.com/aws/eks-anywhere/pkg/curatedpackages"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
	"github.com/spf13/cobra"
	"log"
	"strings"
)

type installPackageOptions struct {
	source      string
	kubeVersion string
	name        string
}

var ipo = &installPackageOptions{}

func init() {
	installCmd.AddCommand(installPackageCommand)
	installPackageCommand.Flags().StringVar(&ipo.source, "source", "", "Location to find curated packages: (cluster, registry)")
	installPackageCommand.Flags().StringVar(&ipo.kubeVersion, "kubeversion", "", "Kubernetes Version of the cluster to be used. Format <major>.<minor>")
	installPackageCommand.Flags().StringVar(&ipo.name, "name", "", "Custom name of the curated package to install")
	if err := installPackageCommand.MarkFlagRequired("source"); err != nil {
		log.Fatalf("Error marking flag as required: %v", err)
	}
}

var installPackageCommand = &cobra.Command{
	Use:          "packages [flags]",
	Aliases:      []string{"package", "packages"},
	Short:        "Install package(s)",
	Long:         "This command is used to Install curated ",
	PreRunE:      preRunPackages,
	SilenceUsage: true,
	RunE:         runInstallPackages(),
}

func runInstallPackages() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		source := strings.ToLower(gepo.source)
		if err := validateSource(source); err != nil {
			return err
		}

		if err := validateKubeVersion(gepo.kubeVersion, source); err != nil {
			return err
		}

		return installPackages(cmd.Context(), ipo, args)
	}
}

func installPackages(ctx context.Context, ipo *installPackageOptions, args []string) error {
	kubeConfig := kubeconfig.FromEnvironment()
	bundle, err := curatedpackages.GetLatestBundle(ctx, kubeConfig, gepo.source, gepo.kubeVersion)
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
