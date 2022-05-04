package curatedpackages

import (
	"context"
	"fmt"
)

type PackageControllerClient struct {
	kubeConfig     string
	ociUri         string
	chartName      string
	chartVersion   string
	chartInstaller ChartInstaller
}

type ChartInstaller interface {
	InstallChartFromName(ctx context.Context, ociURI, kubeConfig, name, version string) error
}

func NewPackageControllerClient(chartInstaller ChartInstaller, kubeConfig, ociUri, chartName, chartVersion string) *PackageControllerClient {
	return &PackageControllerClient{
		kubeConfig:     kubeConfig,
		ociUri:         ociUri,
		chartName:      chartName,
		chartVersion:   chartVersion,
		chartInstaller: chartInstaller,
	}
}

func (pc *PackageControllerClient) InstallController(ctx context.Context) error {
	uri := fmt.Sprintf("%s%s", "oci://", pc.ociUri)
	return pc.chartInstaller.InstallChartFromName(ctx, uri, pc.kubeConfig, pc.chartVersion, pc.chartName)
}
