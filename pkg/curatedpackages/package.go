package curatedpackages

import (
	"context"
	"fmt"
	"github.com/aws/eks-anywhere/pkg/executables"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	api "github.com/aws/eks-anywhere-packages/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

const (
	minWidth        = 16
	tabWidth        = 8
	padding         = 0
	padChar         = '\t'
	flags           = 0
	customName      = "my-"
	kind            = "Package"
	filePermission  = 0o644
	dirPermission   = 0o755
	packageLocation = "curated-packages"
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

func GeneratePackages(bundle *api.PackageBundle, args []string) ([]api.Package, error) {
	packageNameToPackage := getPackageNameToPackage(bundle.Spec.Packages)
	var packages []api.Package
	for _, v := range args {
		bp := packageNameToPackage[strings.ToLower(v)]
		if bp.Name == "" {
			fmt.Println(fmt.Errorf("unknown package %s", v).Error())
			continue
		}
		packageName := customName + strings.ToLower(bp.Name)
		packages = append(packages, convertBundlePackageToPackage(bp, packageName, bundle.APIVersion))
	}
	return packages, nil
}

func WritePackagesToFile(packages []api.Package, d string) error {
	directory := filepath.Join(d, packageLocation)
	if err := os.Mkdir(directory, dirPermission); err != nil {
		return fmt.Errorf("unable to create directory %s", directory)
	}

	for _, p := range packages {
		displayPackage := NewDisplayPackage(p)
		content, err := yaml.Marshal(displayPackage)
		if err != nil {
			fmt.Println(fmt.Errorf("unable to parse package %s %v", p.Name, err).Error())
			continue
		}
		writeToFile(directory, p.Name, content)
	}
	return nil
}

func writeToFile(dir string, packageName string, content []byte) {
	file := filepath.Join(dir, packageName) + ".yaml"
	if err := os.WriteFile(file, content, filePermission); err != nil {
		fmt.Println(fmt.Errorf("unable to write to the file: %s %v", file, err))
	}
}

func GetPackageFromBundle(bundle *api.PackageBundle, packageName string) (*api.BundlePackage, error) {
	packagesInBundle := bundle.Spec.Packages
	pntop := getPackageNameToPackage(packagesInBundle)
	p := pntop[strings.ToLower(packageName)]
	if p.Name != "" {
		return &p, nil
	}
	return nil, fmt.Errorf("package %s not found", packageName)
}

func InstallPackage(ctx context.Context, bp *api.BundlePackage, b *api.PackageBundle, customName string, kubeConfig string) error {
	p := convertBundlePackageToPackage(*bp, customName, b.APIVersion)
	deps, err := newDependencies(ctx)
	if err != nil {
		return fmt.Errorf("unable to initialize executables: %v", err)
	}
	kubectl := deps.Kubectl
	packageYaml, err := yaml.Marshal(p)
	if err != nil {
		return err
	}
	err = kubectl.ApplyResourcesFromBytes(ctx, packageYaml, kubeConfig)
	return err
}

func ApplyResource(ctx context.Context, resource string, fileName string, kubeConfig string) error {
	deps, err := newDependencies(ctx)
	if err != nil {
		return fmt.Errorf("unable to initialize executables: %v", err)
	}
	kubectl := deps.Kubectl
	params := []executables.KubectlOpt{executables.WithKubeconfig(kubeConfig), executables.WithFile(fileName)}
	err = kubectl.ApplyResources(ctx, resource, params...)
	return err
}

func getPackageNameToPackage(packages []api.BundlePackage) map[string]api.BundlePackage {
	pntop := make(map[string]api.BundlePackage)
	for _, p := range packages {
		pntop[strings.ToLower(p.Name)] = p
	}
	return pntop
}

func convertBundlePackageToPackage(bp api.BundlePackage, name string, apiVersion string) api.Package {
	versionToUse := bp.Source.Versions[0]
	p := api.Package{
		ObjectMeta: metav1.ObjectMeta{
			Name:      customName + strings.ToLower(bp.Name),
			Namespace: constants.EksaPackagesName,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       kind,
			APIVersion: apiVersion,
		},
		Spec: api.PackageSpec{
			PackageName:     bp.Name,
			PackageVersion:  versionToUse.Name,
			TargetNamespace: constants.EksaPackagesName,
		},
	}
	return p
}
