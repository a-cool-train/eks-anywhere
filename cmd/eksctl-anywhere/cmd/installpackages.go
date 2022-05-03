package cmd

import (
	"context"
	"fmt"
	"github.com/aws/eks-anywhere/pkg/config"
	"github.com/aws/eks-anywhere/pkg/version"
	"log"

	"github.com/spf13/cobra"

	"github.com/aws/eks-anywhere/pkg/curatedpackages"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
)

type installPackageOptions struct {
	source      curatedpackages.BundleSource
	kubeVersion string
	name        string
	registry    string
}

var ipo = &installPackageOptions{}

func init() {
	installCmd.AddCommand(installPackageCommand)
	installPackageCommand.Flags().Var(&ipo.source, "source", "Location to find curated packages: (cluster, registry)")
	if err := installPackageCommand.MarkFlagRequired("source"); err != nil {
		log.Fatalf("Error marking flag as required: %v", err)
	}
	installPackageCommand.Flags().StringVar(&ipo.kubeVersion, "kubeversion", "", "Kubernetes Version of the cluster to be used. Format <major>.<minor>")
	installPackageCommand.Flags().StringVar(&ipo.name, "name", "", "Custom name of the curated package to install")
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
	deps, err := newDependenciesForPackages(ctx, kubeConfig)
	if err != nil {
		return fmt.Errorf("unable to initialize executables: %v", err)
	}

	bm := curatedpackages.CreateBundleManager(ipo.kubeVersion)
	username, password, err := config.ReadCredentials()
	if err != nil && gpOptions.registry != "" {
		return err
	}
	registry, err := curatedpackages.NewRegistry(deps, ipo.registry, ipo.kubeVersion, username, password)
	if err != nil {
		return err
	}

	b := curatedpackages.NewBundleReader(
		kubeConfig,
		ipo.kubeVersion,
		ipo.source,
		deps.Kubectl,
		bm,
		version.Get(),
		registry,
	)

	bundle, err := b.GetLatestBundle(ctx)
	if err != nil {
		return err
	}

	packages := curatedpackages.NewPackageClient(
		bundle,
		deps.Kubectl,
	)

	p, err := packages.GetPackageFromBundle(args[0])
	if err != nil {
		return err
	}
	err = packages.InstallPackage(ctx, p, ipo.name, kubeConfig)
	if err != nil {
		return err
	}
	return nil
}
