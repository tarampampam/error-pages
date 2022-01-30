package tpl_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tarampampam/error-pages/internal/tpl"
)

func TestHashBytes(t *testing.T) {
	assert.NotEqual(t, tpl.HashBytes([]byte{1}), tpl.HashBytes([]byte{2}))
}

func TestHashStruct(t *testing.T) {
	type s struct {
		S string
		I int
		B bool
	}

	h1, err1 := tpl.HashStruct(s{S: "foo", I: 1, B: false})
	assert.NoError(t, err1)

	h2, err2 := tpl.HashStruct(s{S: "bar", I: 2, B: true})
	assert.NoError(t, err2)

	assert.NotEqual(t, h1, h2)

	type p struct { // no exported fields
		any string
	}

	_, err := tpl.HashStruct(p{any: "foo"})
	assert.Error(t, err)
}
