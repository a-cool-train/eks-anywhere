package curatedpackages

import (
	"context"
	"github.com/aws/eks-anywhere/pkg/dependencies"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
	"path/filepath"
)

func InstallController(ctx context.Context, clusterName string) error {
	kubeConfig := kubeconfig.FromClusterName(clusterName)
	deps, err := newDependenciesWithHelm(ctx, filepath.Dir(kubeConfig))
	if err != nil {
		return err
	}
	helm := deps.Helm
	return helm.InstallChart(ctx, "oci://public.ecr.aws/j0a1m4z9/eks-anywhere-packages", kubeConfig, "0.1.4+ad689eb0f06c6ccfd6f9c3ad130445e2a0e25eb9", "eks-anywhere-packages")
}

func newDependenciesWithHelm(ctx context.Context, paths ...string) (*dependencies.Dependencies, error) {
	return dependencies.NewFactory().
		WithExecutableImage().
		WithExecutableMountDirs(paths...).
		WithExecutableBuilder().
		WithHelm().
		Build(ctx)
}
