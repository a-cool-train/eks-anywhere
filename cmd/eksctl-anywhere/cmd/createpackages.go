package cmd

import (
	"context"
	"log"

	"github.com/spf13/cobra"

	"github.com/aws/eks-anywhere/pkg/curatedpackages"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
)

type createPackageOptions struct {
	fileName string
}

var cpo = &createPackageOptions{}

func init() {
	createCmd.AddCommand(createPackagesCommand)
	createPackagesCommand.Flags().StringVarP(&cpo.fileName, "filename", "f", "", "File with curated packages to create")
	err := createPackagesCommand.MarkFlagRequired("filename")
	if err != nil {
		log.Fatalf("Error marking flag as required: %v", err)
	}
}

var createPackagesCommand = &cobra.Command{
	Use:          "packages",
	Short:        "Create curated packages",
	PreRunE:      preRunPackages,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := createPackages(cmd.Context()); err != nil {
			return err
		}
		return nil
	},
}

func createPackages(ctx context.Context) error {
	kubeConfig := kubeconfig.FromEnvironment()
	err := curatedpackages.ApplyResource(ctx, "create", cpo.fileName, kubeConfig)
	return err
}
