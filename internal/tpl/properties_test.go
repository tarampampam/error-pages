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
		OriginalURI: "aaa",
		Namespace:   "bbb",
		IngressName: "ccc",
		ServiceName: "ddd",
		ServicePort: "eee",
		RequestID:   "fff",
		ForwardedFor:"ggg",
		Host:  		 "hhh",
	}

	r := props.Replaces()

	assert.Equal(t, "foo", r["code"])
	assert.Equal(t, "bar", r["message"])
	assert.Equal(t, "baz", r["description"])
	assert.Equal(t, "aaa", r["original_uri"])
	assert.Equal(t, "bbb", r["namespace"])
	assert.Equal(t, "ccc", r["ingress_name"])
	assert.Equal(t, "ddd", r["service_name"])
	assert.Equal(t, "eee", r["service_port"])
	assert.Equal(t, "fff", r["request_id"])
	assert.Equal(t, "ggg", r["forwarded_for"])
	assert.Equal(t, "hhh", r["host"])

	props.Code, props.Message, props.Description = "", "", ""

	r = props.Replaces()

	assert.Equal(t, "", r["code"])
	assert.Equal(t, "", r["message"])
	assert.Equal(t, "", r["description"])
}
