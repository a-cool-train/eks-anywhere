package oras

import (
	"context"
	"fmt"
	"github.com/aws/eks-anywhere/pkg/curatedpackages"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
	"os"
	"path/filepath"
	"strings"
)

type FileRegistryImporter struct {
	registry           string
	username, password string
	srcFolder          string
}

func NewFileRegistryImporter(registry, username, password, srcFolder string) *FileRegistryImporter {
	return &FileRegistryImporter{
		registry:  registry,
		username:  username,
		password:  password,
		srcFolder: srcFolder,
	}
}

func (fr *FileRegistryImporter) Push(ctx context.Context, bundles *releasev1.Bundles) {
	artifacts := ReadFilesFromBundles(bundles)
	for _, a := range UniqueCharts(artifacts) {
		chartName := filepath.Base(a)
		fileName := ChartFileName(a)
		chartFilepath := filepath.Join(fr.srcFolder, fileName)
		fmt.Println("Path: " + chartFilepath)
		data, err := os.ReadFile(chartFilepath)

		if err != nil {
			fmt.Println(fmt.Errorf("failed reading file: %v", err).Error())
			continue
		}
		ref := fmt.Sprintf("%s/%s", fr.registry, chartName)
		fmt.Println("ref: " + ref)
		curatedpackages.Push(ctx, a, ref, fileName, data)
	}
}

func ChartFileName(chart string) string {
	return strings.Replace(filepath.Base(chart), ":", "-", 1) + ".yaml"
}
