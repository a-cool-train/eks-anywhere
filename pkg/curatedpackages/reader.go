package curatedpackages

import (
	"github.com/aws/eks-anywhere/pkg/manifests"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
	eksdv1 "github.com/aws/eks-distro-build-tooling/release/api/v1alpha1"
)

type PackagesReader struct {
	ManifestReader *manifests.Reader
}

func NewPackageReader(mr *manifests.Reader) *PackagesReader {
	return &PackagesReader{
		ManifestReader: mr,
	}
}

func (r *PackagesReader) ReadBundlesForVersion(version string) (*releasev1.Bundles, error) {
	return r.ManifestReader.ReadBundlesForVersion(version)
}

func (r *PackagesReader) ReadEKSD(eksaVersion, kubeVersion string) (*eksdv1.Release, error) {
	return r.ManifestReader.ReadEKSD(eksaVersion, kubeVersion)
}

func (r *PackagesReader) ReadImages(eksaVersion string) ([]releasev1.Image, error) {
	images, err := r.ManifestReader.ReadImages(eksaVersion)
	return images, err
}

func (r *PackagesReader) ReadImagesFromBundles(b *releasev1.Bundles) ([]releasev1.Image, error) {
	images, err := r.ManifestReader.ReadImagesFromBundles(b)
	return images, err
}

func (r *PackagesReader) ReadCharts(eksaVersion string) ([]releasev1.Image, error) {
	images, err := r.ManifestReader.ReadCharts(eksaVersion)
	return images, err
}

func (r *PackagesReader) ReadChartsFromBundles(b *releasev1.Bundles) []releasev1.Image {
	images := r.ManifestReader.ReadChartsFromBundles(b)
	return images
}
