package curatedpackages

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/go-logr/logr"
	"sigs.k8s.io/yaml"

	api "github.com/aws/eks-anywhere-packages/api/v1alpha1"
	"github.com/aws/eks-anywhere-packages/pkg/artifacts"
	"github.com/aws/eks-anywhere-packages/pkg/bundle"
	"github.com/aws/eks-anywhere-packages/pkg/testutil"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/dependencies"
	"github.com/aws/eks-anywhere/pkg/executables"
)

const (
	RegistryBaseRef = "public.ecr.aws/q0f6t3x4/eksa-package-bundles"
)

func GetLatestBundle(ctx context.Context, kubeConfig string, source BundleSource, kubeVersion string) (*api.PackageBundle, error) {
	switch source {
	case Cluster:
		return getActiveBundleFromCluster(ctx, kubeConfig)
	case Registry:
		return getLatestBundleFromRegistry(ctx, kubeVersion)
	default:
		return nil, fmt.Errorf("unknown source: %q", source)
	}
}

func createBundleManager(kubeVersion string) bundle.Manager {
	versionSplit := strings.Split(kubeVersion, ".")
	major, minor := versionSplit[0], versionSplit[1]
	log := logr.Discard()
	discovery := testutil.NewFakeDiscovery(major, minor)
	puller := artifacts.NewRegistryPuller()
	return bundle.NewBundleManager(log, discovery, puller)
}

func getLatestBundleFromRegistry(ctx context.Context, kubeVersion string) (*api.PackageBundle, error) {
	bm := createBundleManager(kubeVersion)
	return bm.LatestBundle(ctx, RegistryBaseRef)
}

func getActiveBundleFromCluster(ctx context.Context, kubeConfig string) (*api.PackageBundle, error) {
	deps, err := newDependencies(ctx, filepath.Dir(kubeConfig))
	if err != nil {
		return nil, fmt.Errorf("unable to initialize executables: %v", err)
	}
	kubectl := deps.Kubectl

	// Active Bundle is set at the bundle Controller
	bundleController, err := GetActiveController(ctx, kubectl, kubeConfig)
	if err != nil {
		return nil, err
	}
	bundle, err := getPackageBundle(ctx, kubectl, kubeConfig, bundleController.Spec.ActiveBundle)
	if err != nil {
		return nil, err
	}
	return bundle, nil
}

func getPackageBundle(ctx context.Context, kubectl *executables.Kubectl, kubeConfig string, activeBundle string) (*api.PackageBundle, error) {
	params := []executables.KubectlOpt{executables.WithArg("get"), executables.WithArg("packageBundle"), executables.WithOutput("json"), executables.WithKubeconfig(kubeConfig), executables.WithNamespace(constants.EksaPackagesName), executables.WithArg(activeBundle)}
	stdOut, err := kubectl.ApplyResources(ctx, params...)
	if err != nil {
		return nil, err
	}
	obj := &api.PackageBundle{}
	if err := json.Unmarshal(stdOut.Bytes(), obj); err != nil {
		return nil, fmt.Errorf("unmarshaling package bundle: %w", err)
	}
	return obj, nil
}

func GetActiveController(ctx context.Context, kubectl *executables.Kubectl, kubeConfig string) (*api.PackageBundleController, error) {
	params := []executables.KubectlOpt{executables.WithArg("get"), executables.WithArg("packageBundleController"), executables.WithOutput("json"), executables.WithKubeconfig(kubeConfig), executables.WithNamespace(constants.EksaPackagesName), executables.WithArg(bundle.PackageBundleControllerName)}
	stdOut, err := kubectl.ApplyResources(ctx, params...)
	if err != nil {
		return nil, err
	}
	obj := &api.PackageBundleController{}
	if err := json.Unmarshal(stdOut.Bytes(), obj); err != nil {
		return nil, fmt.Errorf("unmarshaling active package bundle controller: %w", err)
	}
	return obj, nil
}

func UpgradeBundle(ctx context.Context, controller *api.PackageBundleController, k *executables.Kubectl, newBundle string, kubeConfig string) error {
	controller.Spec.ActiveBundle = newBundle
	controllerYaml, err := yaml.Marshal(controller)
	if err != nil {
		return err
	}
	params := []executables.KubectlOpt{executables.WithArg("apply"), executables.WithFile("-"), executables.WithKubeconfig(kubeConfig)}
	_, err = k.ApplyResourcesFromBytes(ctx, controllerYaml, params...)
	if err != nil {
		return err
	}
	return nil
}

func newDependencies(ctx context.Context, paths ...string) (*dependencies.Dependencies, error) {
	return dependencies.NewFactory().
		WithExecutableImage(executables.DefaultEksaImage()).
		WithExecutableMountDirs(paths...).
		WithExecutableBuilder().
		WithKubectl().
		Build(ctx)
}
