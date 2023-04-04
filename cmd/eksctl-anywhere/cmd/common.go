package cmd

import (
	"context"
	"fmt"
	"github.com/aws/eks-anywhere/pkg/manifests/bundles"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/dependencies"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
	"github.com/aws/eks-anywhere/pkg/version"
	"github.com/aws/eks-anywhere/release/api/v1alpha1"
)

func getImages(clusterSpecPath, bundlesOverride string) ([]v1alpha1.Image, error) {
	var specOpts []cluster.FileSpecBuilderOpt
	if bundlesOverride != "" {
		specOpts = append(specOpts, cluster.WithOverrideBundlesManifest(bundlesOverride))
	}
	clusterSpec, err := readAndValidateClusterSpec(clusterSpecPath, version.Get(), specOpts...)
	if err != nil {
		return nil, err
	}
	bundle := clusterSpec.VersionsBundle
	images := append(bundle.Images(), clusterSpec.KubeDistroImages()...)
	return images, nil
}

// getKubeconfigPath returns an EKS-A kubeconfig path. The return van be overriden using override
// to give preference to a user specified kubeconfig.
func getKubeconfigPath(clusterName, override string) string {
	if override == "" {
		return kubeconfig.FromClusterName(clusterName)
	}
	return override
}

func NewDependenciesForPackages(ctx context.Context, opts ...PackageOpt) (*dependencies.Dependencies, error) {
	config := New(opts...)
	factory, err := dependencies.NewFactory().WithManifestReader().Build(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize factory %v", err)
	}
	if config.bundlesOverride != "" {
		eksaBundles, err := bundles.Read(factory.ManifestReader, config.bundlesOverride)
		if err != nil {
			return nil, fmt.Errorf("retrieving executable tools image from bundle in dependency factory %v", err)
		}
		// Note: Currently using the first available version of the cli tools
		// This is because the binaries bundled are all the same version hence no compatibility concerns
		// In case, there is a change to this behavior, there might be a need to reassess this item
		image := eksaBundles.Spec.VersionsBundles[0].Eksa.CliTools.VersionedImage()
		factory.UseExecutableImage(image)
	}
	return dependencies.NewFactory().
		WithExecutableMountDirs(config.mountPaths...).
		WithCustomExecutableImage(config.bundlesOverride).
		WithExecutableBuilder().
		WithManifestReader().
		WithKubectl().
		WithHelm(executables.WithInsecure()).
		WithCuratedPackagesRegistry(config.registryName, config.kubeVersion, version.Get()).
		WithPackageControllerClient(config.spec, config.kubeConfig).
		Build(ctx)
}

func UseBundlesOverride(bundlesOverride string) {

}

type PackageOpt func(*PackageConfig)

type PackageConfig struct {
	registryName    string
	kubeVersion     string
	kubeConfig      string
	mountPaths      []string
	spec            *cluster.Spec
	bundlesOverride string
}

func New(options ...PackageOpt) *PackageConfig {
	pc := &PackageConfig{}
	for _, o := range options {
		o(pc)
	}
	return pc
}

func WithRegistryName(registryName string) func(*PackageConfig) {
	return func(config *PackageConfig) {
		config.registryName = registryName
	}
}

func WithKubeVersion(kubeVersion string) func(*PackageConfig) {
	return func(config *PackageConfig) {
		config.kubeVersion = kubeVersion
	}
}

func WithMountPaths(mountPaths ...string) func(*PackageConfig) {
	return func(config *PackageConfig) {
		config.mountPaths = mountPaths
	}
}

func WithClusterSpec(spec *cluster.Spec) func(config *PackageConfig) {
	return func(config *PackageConfig) {
		config.spec = spec
	}
}

func WithKubeConfig(kubeConfig string) func(*PackageConfig) {
	return func(config *PackageConfig) {
		config.kubeConfig = kubeConfig
	}
}

func WithBundlesOverride(bundlesOverride string) func(*PackageConfig) {
	return func(config *PackageConfig) {
		config.bundlesOverride = bundlesOverride
	}
}
