package curatedpackages

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKubeVersionSet(t *testing.T) {
	h := newKVHelper(t)
	h.Parallel()

	type acceptCase struct {
		// name of the test
		name string
		// input from the user
		input string
		// expected result
		expected kubeVersion
	}

	good := kubeVersion{"1", "21"}
	accepts := []acceptCase{
		{"golden path", "1.21", good},
		{"whitespace before", " 1.21", good},
		{"whitespace after", " 1.21 ", good},
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
		{"too many numbers", "1.2.3"},
		{"not enough numbers", "1"},
		{"what?", "1.a"},
		{"negative numbers", "-1"},
		{"negative numbers 2", "1.-1"},
		{"a.b.c", "a.b.c"},
		{"double dash", "--"},
		{"exclamation point", "!"},
		{"something", "something"},
		{"junk", "junk"},
		{"bundleSource", "cluster"},
		{"random space in the middle", "1. 21"},
	}
	for _, testcase := range rejects {
		h.runRejectCase(testcase.name, testcase.input)
	}
}

//
// Helpers
//

type kvHelper struct{ *testing.T }

func newKVHelper(t *testing.T) *kvHelper {
	return &kvHelper{T: t}
}

func (h *kvHelper) runAcceptCase(name, input string, expected kubeVersion) {
	h.Helper()
	h.Run("accepts "+name, func(t *testing.T) {
		h.assertAccepts(assert.New(t), input, expected)
	})
}

func (h *kvHelper) assertAccepts(t *assert.Assertions, input string,
	expected kubeVersion) {

	kv := kubeVersion{}
	if t.NoError(kv.Set(input)) {
		t.Equal(expected, kv)
	}
}

func (h *kvHelper) runRejectCase(name, input string) {
	h.Helper()
	h.Run("rejects "+name, func(t *testing.T) {
		h.assertRejects(assert.New(t), input)
	})
}

func (h *kvHelper) assertRejects(t *assert.Assertions, input string) {
	kv := kubeVersion{}
	t.Error(kv.Set(input))
}
