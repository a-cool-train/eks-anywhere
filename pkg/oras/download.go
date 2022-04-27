package oras

import (
	"bytes"
	"context"
	"fmt"
	"github.com/aws/eks-anywhere-packages/pkg/artifacts"
	"github.com/aws/eks-anywhere/pkg/types"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type BundleDownloader struct {
	dstFolder string
}

func NewBundleDownloader(dstFolder string) *BundleDownloader {
	return &BundleDownloader{
		dstFolder: dstFolder,
	}
}

func (bd *BundleDownloader) Pull(ctx context.Context, arts ...string) error {
	puller := artifacts.NewRegistryPuller()

	for _, a := range uniqueCharts(arts) {
		data, err := puller.Pull(ctx, a)
		if err != nil {
			return fmt.Errorf("unable to pull artifacts %v", err)
		}
		if len(bytes.TrimSpace(data)) == 0 {
			return fmt.Errorf("latest package bundle artifact is empty")
		}
		writeToFile(bd.dstFolder, "eks-anywhere-packages-bundles", data)
	}
	return nil
}

func uniqueCharts(charts []string) []string {
	c := types.SliceToLookup(charts).ToSlice()
	// TODO: maybe optimize this, avoiding the sort and just following the same order as the original slice
	sort.Strings(c)
	return c
}

func writeToFile(dir string, packageName string, content []byte) error {
	file := filepath.Join(dir, packageName) + ".yaml"
	if err := os.WriteFile(file, content, 0o644); err != nil {
		return fmt.Errorf("unable to write to the file: %s %v", file, err)
	}
	return nil
}

func (bd *BundleDownloader) ReadFilesFromBundles(bundles *releasev1.Bundles) []string {
	var files []string
	for _, vb := range bundles.Spec.VersionsBundles {
		files = append(files, getPackageBundle(vb))
	}
	return files
}

func getPackageBundle(vb releasev1.VersionsBundle) string {
	packageController := vb.PackageController
	// Use package controller registry to fetch packageBundles.
	// Format of controller image is: <uri>/<env_type>/<repository_name>
	controllerImage := strings.Split(packageController.Controller.Image(), "/")
	registryBaseRef := fmt.Sprintf("%s/%s/%s", controllerImage[0], controllerImage[1], "eks-anywhere-packages-bundles")
	fmt.Println("Registry Base Ref: " + registryBaseRef)
	return registryBaseRef
}
