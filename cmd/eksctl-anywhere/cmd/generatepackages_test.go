package cmd

import (
	"context"
	"os/exec"
	"testing"
)

func BenchmarkGeneratePackageGoLibrary(b *testing.B) {
	ctx := context.Background()

	for n := 0; n < b.N; n++ {
		GoLibraryExecution(ctx)
	}
}

func BenchmarkGeneratePackageKubectl(b *testing.B) {
	ctx := context.Background()

	for n := 0; n < b.N; n++ {
		KubectlContainerExecution(ctx)
	}
}

func BenchmarkGeneratePackageKubectlMRTOOLSDisable(b *testing.B) {
	ctx := context.Background()

	for n := 0; n < b.N; n++ {
		KubectlContainerExecutionMRToolsDisable(ctx)
	}
}

func BenchmarkGeneratePackageKubectlNative(b *testing.B) {
	for n := 0; n < b.N; n++ {
		exec.Command("/Users/acool/Desktop/Amazon/eks-anywhere-cluster/kubectl.sh").Output()
	}
}
