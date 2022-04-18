package curatedpackages

import (
	"context"
	"fmt"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/dependencies"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
	"path/filepath"
)

func InstallController(ctx context.Context, cs *cluster.Spec) error {
	kubeConfig := kubeconfig.FromClusterName(cs.Cluster.Name)
	deps, err := newDependenciesWithHelm(ctx, filepath.Dir(kubeConfig))
	if err != nil {
		return err
	}
	helm := deps.Helm
	helmChart := cs.VersionsBundle.VersionsBundle.PackageController.HelmChart
	uri := fmt.Sprintf("%s%s", "oci://", helmChart.Image())
	return helm.InstallChart(ctx, uri, kubeConfig, helmChart.Tag(), helmChart.Name)
}

func newDependenciesWithHelm(ctx context.Context, paths ...string) (*dependencies.Dependencies, error) {
	return dependencies.NewFactory().
		WithExecutableImage().
		WithExecutableMountDirs(paths...).
		WithExecutableBuilder().
		WithHelm().
		Build(ctx)
}
