package curatedpackages

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/pflag"
)

// KubeVersion specifies a kubernetes cluster version.
//
// It implements pflag.Value, and expects the caller to call Set to change its
// value from it's default invalid state, to a valid state.
type KubeVersion struct {
	kubeVersion // wrapped so that this value is hidden
}

var _ pflag.Value = (*KubeVersion)(nil)

type kubeVersion struct {
	// Major version for kubernetes.
	Major string
	// Minor version for kubernetes.
	Minor string
}

// Set parses the user's input into a kubeVersion.
func (kv *kubeVersion) Set(s string) error {
	var err error

	pieces := strings.Split(strings.TrimSpace(s), ".")
	if len(pieces) != 2 {
		return fmt.Errorf("kubernetes version must be specified as <major>.<minor>, e.g. 1.22")
	}

	// We don't actually want an integer value, but we want it to be a valid
	// string representation of an integer.
	major, err := strconv.Atoi(pieces[0])
	if err != nil || major < 1 {
		return fmt.Errorf("invalid kubernetes major version %q: %w", pieces[0], err)
	}

	// Same as for the major version.
	minor, err := strconv.Atoi(pieces[1])
	if err != nil || minor < 0 {
		return fmt.Errorf("invalid kubernetes minor version %q: %w", pieces[1], err)
	}

	kv.Major = pieces[0]
	kv.Minor = pieces[1]

	return nil
}

func (kv kubeVersion) String() string {
	return fmt.Sprintf("%s.%s", kv.Major, kv.Minor)
}

func (kv kubeVersion) Type() string {
	return "kubeVersion"
}
