package v1alpha1

import (
	"context"
	"fmt"
	api "github.com/aws/eks-anywhere-packages/api/v1alpha1"
	"github.com/aws/eks-anywhere-packages/pkg/artifacts"
	"github.com/aws/eks-anywhere-packages/pkg/bundle"
	"github.com/aws/eks-anywhere-packages/pkg/testutil"
	"github.com/go-logr/logr"
	"strings"
)

func (vb *VersionsBundle) Manifests() map[string][]*string {
	return map[string][]*string{
		"cluster-api-provider-aws": {
			&vb.Aws.Components.URI,
			&vb.Aws.ClusterTemplate.URI,
			&vb.Aws.Metadata.URI,
		},
		"core-cluster-api": {
			&vb.ClusterAPI.Components.URI,
			&vb.ClusterAPI.Metadata.URI,
		},
		"capi-kubeadm-bootstrap": {
			&vb.Bootstrap.Components.URI,
			&vb.Bootstrap.Metadata.URI,
		},
		"capi-kubeadm-control-plane": {
			&vb.ControlPlane.Components.URI,
			&vb.ControlPlane.Metadata.URI,
		},
		"cert-manager": {
			&vb.CertManager.Manifest.URI,
		},
		"cluster-api-provider-docker": {
			&vb.Docker.Components.URI,
			&vb.Docker.ClusterTemplate.URI,
			&vb.Docker.Metadata.URI,
		},
		"cluster-api-provider-vsphere": {
			&vb.VSphere.Components.URI,
			&vb.VSphere.ClusterTemplate.URI,
			&vb.VSphere.Metadata.URI,
		},
		"cluster-api-provider-cloudstack": {
			&vb.CloudStack.Components.URI,
			&vb.CloudStack.Metadata.URI,
		},
		"cluster-api-provider-tinkerbell": {
			&vb.Tinkerbell.Components.URI,
			&vb.Tinkerbell.ClusterTemplate.URI,
			&vb.Tinkerbell.Metadata.URI,
		},
		"cluster-api-provider-snow": {
			&vb.Snow.Components.URI,
			&vb.Snow.Metadata.URI,
		},
		"cilium": {
			&vb.Cilium.Manifest.URI,
		},
		"kindnetd": {
			&vb.Kindnetd.Manifest.URI,
		},
		"eks-anywhere-cluster-controller": {
			&vb.Eksa.Components.URI,
		},
		"etcdadm-bootstrap-provider": {
			&vb.ExternalEtcdBootstrap.Components.URI,
			&vb.ExternalEtcdBootstrap.Metadata.URI,
		},
		"etcdadm-controller": {
			&vb.ExternalEtcdController.Components.URI,
			&vb.ExternalEtcdController.Metadata.URI,
		},
		"eks-distro": {
			&vb.EksD.Components,
			&vb.EksD.EksDReleaseUrl,
		},
	}
}

func (vb *VersionsBundle) Ovas() []Archive {
	return []Archive{
		vb.EksD.Ova.Bottlerocket.Archive,
		vb.EksD.Ova.Ubuntu.Archive,
	}
}

func (vb *VersionsBundle) CloudStackImages() []Image {
	return []Image{
		vb.CloudStack.ClusterAPIController,
		vb.CloudStack.KubeVip,
	}
}

func (vb *VersionsBundle) VsphereImages() []Image {
	return []Image{
		vb.VSphere.ClusterAPIController,
		vb.VSphere.Driver,
		vb.VSphere.KubeProxy,
		vb.VSphere.KubeVip,
		vb.VSphere.Manager,
		vb.VSphere.Syncer,
	}
}

func (vb *VersionsBundle) DockerImages() []Image {
	return []Image{
		vb.Docker.KubeProxy,
		vb.Docker.Manager,
	}
}

func (vb *VersionsBundle) SnowImages() []Image {
	return []Image{
		vb.Snow.KubeVip,
		vb.Snow.Manager,
	}
}

func (vb *VersionsBundle) SharedImages() []Image {
	return []Image{
		vb.Bootstrap.Controller,
		vb.Bootstrap.KubeProxy,
		vb.BottleRocketBootstrap.Bootstrap,
		vb.BottleRocketAdmin.Admin,
		vb.CertManager.Acmesolver,
		vb.CertManager.Cainjector,
		vb.CertManager.Controller,
		vb.CertManager.Webhook,
		vb.Cilium.Cilium,
		vb.Cilium.Operator,
		vb.ClusterAPI.Controller,
		vb.ClusterAPI.KubeProxy,
		vb.ControlPlane.Controller,
		vb.ControlPlane.KubeProxy,
		vb.EksD.KindNode,
		vb.Eksa.CliTools,
		vb.Eksa.ClusterController,
		vb.Flux.HelmController,
		vb.Flux.KustomizeController,
		vb.Flux.NotificationController,
		vb.Flux.SourceController,
		vb.ExternalEtcdBootstrap.Controller,
		vb.ExternalEtcdBootstrap.KubeProxy,
		vb.ExternalEtcdController.Controller,
		vb.ExternalEtcdController.KubeProxy,
		vb.Haproxy.Image,
	}
}

func (vb *VersionsBundle) Images() []Image {
	groupedImages := [][]Image{
		vb.SharedImages(),
		vb.DockerImages(),
		vb.VsphereImages(),
		vb.CloudStackImages(),
		vb.SnowImages(),
	}

	size := 0
	for _, g := range groupedImages {
		size += len(g)
	}

	images := make([]Image, 0, size)
	for _, g := range groupedImages {
		images = append(images, g...)
	}

	return images
}

func (vb *VersionsBundle) Charts() map[string]*Image {
	imagesMap := make(map[string]*Image)
	imagesMap["cilium"] = &vb.Cilium.HelmChart

	//packageController := vb.PackageController.Controller
	//bundle := Image{
	//	Name:        "packages-bundle-image",
	//	Description: "curated packages bundle image",
	//	OS:          packageController.OS,
	//	OSName:      packageController.OSName,
	//	URI:         getPackageBundleUri(packageController, vb.KubeVersion),
	//}
	//imagesMap[bundle.Name] = &bundle

	images, err := vb.CuratedPackagesImages()
	if err != nil {
		fmt.Println(err)
	}
	for _, image := range images {
		icopy := image
		imagesMap[image.Name] = &icopy
	}
	return imagesMap
}

func (vb *VersionsBundle) CuratedPackagesImages() ([]Image, error) {
	packageController := vb.PackageController.Controller

	packageBundle, err := vb.getPackageBundle()
	if err != nil {
		return []Image{}, fmt.Errorf("unable to parse: %v", err)
	}
	var images []Image
	for _, p := range packageBundle.Spec.Packages {
		pI := Image{
			Name:        p.Name,
			Description: p.Name,
			OS:          packageController.OS,
			OSName:      packageController.OSName,
			URI:         p.Source.Registry + "/" + p.Source.Repository + ":" + p.Source.Versions[0].Name,
		}
		images = append(images, pI)
	}
	return images, nil
}

func (vb *VersionsBundle) PackagesControllerImage() []Image {
	return []Image{
		vb.PackageController.Controller,
	}
}

func (vb *VersionsBundle) getPackageBundle() (*api.PackageBundle, error) {
	bm := createBundleManager(vb.KubeVersion)
	packageController := vb.PackageController
	// Use package controller registry to fetch packageBundles.
	// Format of controller image is: <uri>/<env_type>/<repository_name>
	controllerImage := strings.Split(packageController.Controller.Image(), "/")
	registryBaseRef := fmt.Sprintf("%s/%s/%s", controllerImage[0], controllerImage[1], "eks-anywhere-packages-bundles")
	return bm.LatestBundle(context.Background(), registryBaseRef)
}

func createBundleManager(kubeVersion string) bundle.Manager {
	versionSplit := strings.Split(kubeVersion, ".")
	major, minor := versionSplit[0], versionSplit[1]
	log := logr.Discard()
	discovery := testutil.NewFakeDiscovery(major, minor)
	puller := artifacts.NewRegistryPuller()
	return bundle.NewBundleManager(log, discovery, puller)
}
