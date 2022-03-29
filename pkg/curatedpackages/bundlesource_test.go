package curatedpackages

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBundleSourceSet(t *testing.T) {
	h := newBSHelper(t)
	h.Parallel()

	type acceptCase struct {
		// name of the test
		name string
		// input from the user
		input string
		// expected result
		expected bundleSource
	}

	accepts := []acceptCase{
		{"golden path", "cluster", Cluster},
		{"case insensitivity", "cLuSTEr", Cluster},
		{"golden path", "registry", Registry},
		{"case insensitivity #2", "Registry", Registry},
		{"whitespace before", " registry", Registry},
		{"whitespace after", "registry ", Registry},
	}
	for _, testcase := range accepts {
		h.runAcceptCase(testcase.name, testcase.input, testcase.expected)
	}

	type rejectCase struct {
		// name of the test
		name string
		// input from the user
		input string
	}

	rejects := []rejectCase{
		{"empty", ""},
		{"double dash", "--"},
		{"exclamation point", "!"},
		{"something", "something"},
		{"junk", "junk"},
		{"kubeVersion", "1.21"},
		{"random space in the middle", "reg istry"},
	}
	for _, testcase := range rejects {
		h.runRejectCase(testcase.name, testcase.input)
	}
}

//
// Helpers
//

type bsHelper struct{ *testing.T }

func newBSHelper(t *testing.T) *bsHelper {
	return &bsHelper{T: t}
}

func (h *bsHelper) runAcceptCase(name, input string, expected bundleSource) {
	h.Helper()
	h.Run("accepts "+name, func(t *testing.T) {
		h.assertAccepts(assert.New(t), input, expected)
	})
}

func (h *bsHelper) assertAccepts(t *assert.Assertions,
	input string, expected bundleSource) {

	bs := bundleSource("")
	if t.NoError(bs.Set(input)) {
		t.Equal(expected, bs)
	}
}

func (h *bsHelper) runRejectCase(name, input string) {
	h.Helper()
	h.Run("rejects "+name, func(t *testing.T) {
		h.assertRejects(assert.New(t), input)
	})
}

func (h *bsHelper) assertRejects(t *assert.Assertions, input string) {
	bs := bundleSource("")
	t.Error(bs.Set(input))
}
