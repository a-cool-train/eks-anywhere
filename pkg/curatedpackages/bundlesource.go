package curatedpackages

import (
	"fmt"
	"strings"

	"github.com/spf13/pflag"
)

// BundleSource specifies a source for pulling package bundle information.
//
// It implements pflag.Value, and expects the caller to call Set to change its
// value from it's default invalid state, to a valid state.
type BundleSource struct {
	bundleSource // wrapped so that this value is hidden
}

var _ pflag.Value = (*BundleSource)(nil)

type bundleSource string

const (
	// Cluster indicates that bundles should be retrieved from a cluster.
	Cluster bundleSource = "cluster"
	// Registry indicates that bundles should be retrieved from a registry.
	Registry bundleSource = "registry"
)

func (b bundleSource) String() string {
	return string(b)
}

// Set parses user input into a valid bundle source.
func (b *bundleSource) Set(s string) error {
	src := bundleSource(strings.ToLower(strings.TrimSpace(s)))
	switch src {
	case Cluster, Registry:
		*b = src
	default:
		return fmt.Errorf("unknown bundle source: %q", s)
	}
	return nil
}

func (b bundleSource) Type() string {
	return "BundleSource"
}
