package template_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gh.tarampamp.am/error-pages/internal/template"
)

func TestProps_Values(t *testing.T) {
	t.Parallel()

	assert.Equal(t, template.Props{
		Code:               "a",
		Message:            "b",
		Description:        "c",
		OriginalURI:        "d",
		Namespace:          "e",
		IngressName:        "f",
		ServiceName:        "g",
		ServicePort:        "h",
		RequestID:          "i",
		ForwardedFor:       "j",
		L10nDisabled:       true,
		ShowRequestDetails: false,
	}.Values(), map[string]any{
		"code":          "a",
		"message":       "b",
		"description":   "c",
		"original_uri":  "d",
		"namespace":     "e",
		"ingress_name":  "f",
		"service_name":  "g",
		"service_port":  "h",
		"request_id":    "i",
		"forwarded_for": "j",
		"host":          "", // empty because it's not set
		"l10n_disabled": true,
		"show_details":  false,
	})
}
