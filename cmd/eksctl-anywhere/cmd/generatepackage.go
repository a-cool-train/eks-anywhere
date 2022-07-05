package cmd

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"

	packagesv1 "github.com/aws/eks-anywhere-packages/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/curatedpackages"
	"github.com/aws/eks-anywhere/pkg/version"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	schemav2 "k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	schemav1 "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type generatePackageOptions struct {
	source         curatedpackages.BundleSource
	kubeVersion    string
	registry       string
	useLibrary     bool
	mrToolsDisable bool
}

var gpOptions = &generatePackageOptions{}

func init() {
	generateCmd.AddCommand(generatePackageCommand)
	generatePackageCommand.Flags().Var(&gpOptions.source, "source", "Location to find curated packages: (cluster, registry)")
	if err := generatePackageCommand.MarkFlagRequired("source"); err != nil {
		log.Fatalf("Error marking flag as required: %v", err)
	}
	generatePackageCommand.Flags().StringVar(&gpOptions.kubeVersion, "kube-version", "", "Kubernetes Version of the cluster to be used. Format <major>.<minor>")
	generatePackageCommand.Flags().StringVar(&gpOptions.registry, "registry", "", "Used to specify an alternative registry for package generation")
	generatePackageCommand.Flags().BoolVarP(&gpOptions.useLibrary, "use-library", "u", false, "Specifies whether to use library or container")
	generatePackageCommand.Flags().BoolVarP(&gpOptions.mrToolsDisable, "mr-tools-disable", "m", false, "Specifies whether to use mr tools or not")
}

var generatePackageCommand = &cobra.Command{
	Use:          "packages [flags]",
	Aliases:      []string{"package", "packages"},
	Short:        "Generate package(s) configuration",
	Long:         "Generates Kubernetes configuration files for curated packages",
	PreRunE:      preRunPackages,
	SilenceUsage: true,
	RunE:         runGeneratePackages,
	Args:         cobra.MinimumNArgs(1),
}

func runGeneratePackages(cmd *cobra.Command, args []string) error {
	if err := curatedpackages.ValidateKubeVersion(gpOptions.kubeVersion, gpOptions.source); err != nil {
		return err
	}
	return generatePackages(cmd.Context(), args, gpOptions.useLibrary, gpOptions.mrToolsDisable)
}

func generatePackages(ctx context.Context, args []string, useLibrary bool, mrToolsDisable bool) error {
	//kubeConfig := kubeconfig.FromEnvironment()

	if useLibrary {
		return GoLibraryExecution(ctx)
	}

	if mrToolsDisable {
		KubectlContainerExecutionMRToolsDisable(ctx)
	}
	return KubectlContainerExecution(ctx)
}

func KubectlContainerExecutionMRToolsDisable(ctx context.Context) error {
	os.Setenv("MR_TOOLS_DISABLE", "true")
	err := KubectlContainerExecution(ctx)
	return err
}

func KubectlContainerExecution(ctx context.Context) error {
	kubeConfig := "/Users/acool/Desktop/Amazon/eks-anywhere/mgmt/mgmt-eks-a-cluster.kubeconfig"
	args := []string{"harbor", "hello-eks-anywhere"}
	bm := curatedpackages.CreateBundleManager("1.22")
	deps, err := NewDependenciesForPackages(ctx, WithRegistryName(""), WithKubeVersion("1.22"), WithMountPaths(kubeConfig))
	if err != nil {
		return fmt.Errorf("unable to initialize executables: %v", err)
	}

	b := curatedpackages.NewBundleReader(
		kubeConfig,
		"1.22",
		"cluster",
		deps.Kubectl,
		bm,
		version.Get(),
		deps.BundleRegistry,
	)

	bundle, err := b.GetLatestBundle(ctx)
	if err != nil {
		return err
	}

	packageClient := curatedpackages.NewPackageClient(
		curatedpackages.WithKubectl(deps.Kubectl),
		curatedpackages.WithBundle(bundle),
		curatedpackages.WithCustomPackages(args),
	)
	packages, err := packageClient.GeneratePackages()
	if err != nil {
		return err
	}
	if err = packageClient.WritePackagesToStdOut(packages); err != nil {
		return err
	}
	return nil
}

func GoLibraryExecution(ctx context.Context) error {
	kubeConfig := "/Users/acool/Desktop/Amazon/eks-anywhere/mgmt/mgmt-eks-a-cluster.kubeconfig"
	args := []string{"harbor", "hello-eks-anywhere"}
	result, err := useGoLibrary(ctx, kubeConfig)
	if err != nil {
		return err
	}
	packageClient := curatedpackages.NewPackageClient(
		curatedpackages.WithBundle(result),
		curatedpackages.WithCustomPackages(args),
	)
	_, err = packageClient.GeneratePackages()
	if err != nil {
		return err
	}
	//if err = packageClient.WritePackagesToStdOut(packages); err != nil {
	//	return err
	//}

	return nil
}

func useGoLibrary(ctx context.Context, kubeConfig string) (*packagesv1.PackageBundle, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to parse config %v", err)
	}

	AddToScheme(schemav1.Scheme)
	crdConfig := *config
	crdConfig.ContentConfig.GroupVersion = &schemav2.GroupVersion{Group: GroupName, Version: GroupVersion}
	crdConfig.APIPath = "/apis"
	crdConfig.NegotiatedSerializer = serializer.NewCodecFactory(schemav1.Scheme)
	crdConfig.UserAgent = rest.DefaultKubernetesUserAgent()

	exampleRestClient, err := rest.UnversionedRESTClientFor(&crdConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to parse: %v", err)
	}

	//params := []string{"get", "packageBundleController", "-o", "json", "--kubeconfig", b.kubeConfig, "--namespace", constants.EksaPackagesName, packagesv1.PackageBundleControllerName}
	ctrl := packagesv1.PackageBundleController{}
	err = exampleRestClient.Get().Resource("PackageBundleControllers").Namespace(constants.EksaPackagesName).Suffix(packagesv1.PackageBundleControllerName).Param("-o", "json").Do(ctx).Into(&ctrl)
	if err != nil {
		return nil, fmt.Errorf("couldn't find controller: %v", err)
	}
	result := packagesv1.PackageBundle{}
	err = exampleRestClient.Get().Resource("packagebundles").Namespace(constants.EksaPackagesName).Suffix(ctrl.Spec.ActiveBundle).Do(ctx).Into(&result)

	if err != nil {
		return nil, fmt.Errorf("problem parsing packages %v", err)
	}
	return &result, nil
}

const GroupName = "packages.eks.amazonaws.com"
const GroupVersion = "v1alpha1"

var SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)

var AddToScheme = SchemeBuilder.AddToScheme

var SchemeGroupVersion = schemav2.GroupVersion{Group: GroupName, Version: GroupVersion}

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&packagesv1.PackageBundle{},
		&packagesv1.PackageBundleList{},
		&packagesv1.PackageBundleController{},
	)

	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}
