package curatedpackages

import (
	"context"
	"fmt"
	"github.com/aws/eks-anywhere/pkg/validations"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	api "github.com/aws/eks-anywhere-packages/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/executables"
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
	fmt.Fprintf(w, "%s\t%s\t \n", "Package", "Version(s)")
	fmt.Fprintf(w, "%s\t%s\t \n", "-------", "----------")
	for _, pkg := range packages {
		versions := convertBundleVersionToPackageVersion(pkg.Source.Versions)
		fmt.Fprintf(w, "%s\t%s\t \n", pkg.Name, strings.Join(versions, ","))
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
	if !validations.FileExists(directory) {
		if err := os.Mkdir(directory, dirPermission); err != nil {
			return fmt.Errorf("unable to create directory %s", directory)
		}
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
	deps, err := newDependencies(ctx, kubeConfig)
	if err != nil {
		return fmt.Errorf("unable to initialize executables: %v", err)
	}
	kubectl := deps.Kubectl
	packageYaml, err := yaml.Marshal(p)
	if err != nil {
		return err
	}
	err = kubectl.ApplyResourcesFromBytes(ctx, packageYaml)
	return err
}

func ApplyResource(ctx context.Context, resource string, fileName string, kubeConfig string) error {
	deps, err := newDependencies(ctx, kubeConfig)
	if err != nil {
		return fmt.Errorf("unable to initialize executables: %v", err)
	}
	kubectl := deps.Kubectl
	params := []executables.KubectlOpt{executables.WithArg(resource), executables.WithKubeconfig(kubeConfig), executables.WithFile(fileName)}
	_, err = kubectl.ApplyResources(ctx, params...)
	return err
}

func DeletePackages(ctx context.Context, args []string, kubeConfig string) error {
	deps, err := newDependencies(ctx, kubeConfig)
	if err != nil {
		return fmt.Errorf("unable to initialize executables: %v", err)
	}
	kubectl := deps.Kubectl
	params := []executables.KubectlOpt{executables.WithArg("delete"), executables.WithArg("packages"), executables.WithKubeconfig(kubeConfig), executables.WithArgs(args), executables.WithNamespace(constants.EksaPackagesName)}
	_, err = kubectl.ApplyResources(ctx, params...)
	return err
}

func DescribePackages(ctx context.Context, args []string, kubeConfig string) error {
	deps, err := newDependencies(ctx, kubeConfig)
	if err != nil {
		return fmt.Errorf("unable to initialize executables: %v", err)
	}
	kubectl := deps.Kubectl
	params := []executables.KubectlOpt{executables.WithArg("describe"), executables.WithArg("packages"), executables.WithKubeconfig(kubeConfig), executables.WithArgs(args), executables.WithNamespace(constants.EksaPackagesName)}
	stdOut, err := kubectl.ApplyResources(ctx, params...)
	if err != nil {
		fmt.Print(&stdOut)
		return fmt.Errorf("kubectl execution failure: \n%v", err)
	}
	if len(stdOut.Bytes()) == 0 {
		return fmt.Errorf("no resources found in %s", constants.EksaPackagesName)
	}
	fmt.Println(&stdOut)
	return nil
}

func getPackageNameToPackage(packages []api.BundlePackage) map[string]api.BundlePackage {
	pntop := make(map[string]api.BundlePackage)
	for _, p := range packages {
		pntop[strings.ToLower(p.Name)] = p
	}
	return pntop
}

func convertBundlePackageToPackage(bp api.BundlePackage, name string, apiVersion string) api.Package {
	p := api.Package{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: constants.EksaPackagesName,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       kind,
			APIVersion: apiVersion,
		},
		Spec: api.PackageSpec{
			PackageName:     bp.Name,
			TargetNamespace: constants.EksaPackagesName,
		},
	}
	return p
}
