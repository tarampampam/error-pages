package tpl_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tarampampam/error-pages/internal/tpl"
)

func TestProperties_Replaces(t *testing.T) {
	props := tpl.Properties{
		Code:        "foo",
		Message:     "bar",
		Description: "baz",
	}

	r := props.Replaces()

	assert.Equal(t, "foo", r["code"])
	assert.Equal(t, "bar", r["message"])
	assert.Equal(t, "baz", r["description"])

	props.Code, props.Message, props.Description = "", "", ""

	r = props.Replaces()

	assert.Equal(t, "", r["code"])
	assert.Equal(t, "", r["message"])
	assert.Equal(t, "", r["description"])
}
