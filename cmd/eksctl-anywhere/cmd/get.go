package cmd

import (
	"context"
	"fmt"
	"log"
	"path"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/dependencies"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/features"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get resources",
	Long:  "Use eksctl anywhere get to display one or many resources",
}

func init() {
	rootCmd.AddCommand(getCmd)
}

func preRunPackages(cmd *cobra.Command, args []string) error {
	if !features.IsActive(features.CuratedPackagesSupport()) {
		return fmt.Errorf("this command is currently not supported")
	}
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		if err := viper.BindPFlag(flag.Name, flag); err != nil {
			log.Fatalf("Error initializing flags: %v", err)
		}
	})
	return nil
}

func getResources(ctx context.Context, resourceType string, output string, args []string) error {
	kubeConfig := kubeconfig.FromEnvironment()

	deps, err := newDependencies(ctx, kubeConfig)
	if err != nil {
		return fmt.Errorf("unable to initialize executables: %v", err)
	}
	kubectl := deps.Kubectl

	params := []executables.KubectlOpt{executables.WithArg("get"), executables.WithArg(resourceType), executables.WithKubeconfig(kubeConfig), executables.WithArgs(args), executables.WithNamespace(constants.EksaPackagesName)}
	if output != "" {
		params = append(params, executables.WithOutput(output))
	}
	stdOut, err := kubectl.ApplyResources(ctx, params...)
	if err != nil {
		fmt.Print(&stdOut)
		return fmt.Errorf("kubectl execution failure: \n%v", err)
	}
	if len(stdOut.Bytes()) == 0 {
		return fmt.Errorf("No resources found in %v namespace\n", constants.EksaPackagesName)
	}
	fmt.Print(&stdOut)
	return nil
}

func newDependencies(ctx context.Context, kubeConfig string) (*dependencies.Dependencies, error) {
	return dependencies.NewFactory().
		WithExecutableImage(executables.DefaultEksaImage()).
		WithExecutableMountDirs(path.Dir(kubeConfig)).
		WithExecutableBuilder().
		WithKubectl().
		Build(ctx)
}
