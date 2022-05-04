package oras

import (
	"context"
	"fmt"
	"github.com/aws/eks-anywhere/pkg/curatedpackages"
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

func (bd *BundleDownloader) SaveManifests(ctx context.Context, bundles *releasev1.Bundles) {
	artifacts := bd.ReadFilesFromBundles(bundles)
	for _, a := range uniqueCharts(artifacts) {
		data, err := curatedpackages.Pull(ctx, a)
		if err != nil {
			fmt.Sprintf("unable to download bundle %v", err)
			continue
		}
		bundleName := strings.Replace(filepath.Base(a), ":", "-", 1)
		writeToFile(bd.dstFolder, bundleName, data)
	}
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
		files = append(files, curatedpackages.GetPackageBundle(vb))
	}
	return files
}
