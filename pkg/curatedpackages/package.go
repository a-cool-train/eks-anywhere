package curatedpackages

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	api "github.com/aws/eks-anywhere-packages/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
)

const (
	minWidth = 16
	tabWidth = 8
	padding  = 0
	padChar  = '\t'
	flags    = 0
)

func DisplayPackages(packages []api.BundlePackage) {
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, minWidth, tabWidth, padding, padChar, flags)
	defer w.Flush()
	fmt.Fprintf(w, "\n %s\t%s\t", "Package", "Version(s)")
	fmt.Fprintf(w, "\n %s\t%s\t", "----", "----")
	for _, pkg := range packages {
		versions := convertBundleVersionToPackageVersion(pkg.Source.Versions)
		fmt.Fprintf(w, "\n %s\t%s\t", pkg.Name, strings.Join(versions, ","))
	}
}

func convertBundleVersionToPackageVersion(bundleVersions []api.SourceVersion) []string {
	var versions []string
	for _, v := range bundleVersions {
		versions = append(versions, v.Name)
	}
	return versions
}

func ListPackages(ctx context.Context, src BundleSource, kv KubeVersion) error {
	kubeConfig := kubeconfig.FromEnvironment()
	bundle, err := getLatestBundle(ctx, kubeConfig, src.bundleSource, kv.kubeVersion)
	if err != nil {
		return err
	}
	packages := bundle.Spec.Packages
	DisplayPackages(packages)
	return nil
}
