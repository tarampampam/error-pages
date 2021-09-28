package tpl_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tarampampam/error-pages/internal/tpl"
)

func ExampleReplace() {
	var in = []byte("{{ code }}: {{message}} ({{ description }})")

	fmt.Println(string(tpl.Replace(in, tpl.Replaces{
		Code:        "400",
		Message:     "Bad Request",
		Description: "The server did not understand the request",
	})))

	// Output:
	// 400: Bad Request (The server did not understand the request)
}

func TestReplace(t *testing.T) {
	for name, tt := range map[string]struct {
		giveIn     []byte
		giveRe     tpl.Replaces
		wantResult []byte
	}{
		"common": {
			giveIn: []byte("-- {{ code }} {{code}} __ {{message}} {{ description }} "),
			giveRe: tpl.Replaces{
				Code:        "123",
				Message:     "message",
				Description: "desc",
			},
			wantResult: []byte("-- 123 123 __ message desc "),
		},
		"alpha and underline in the code": {
			giveIn: []byte("\t{{ code }}\t"),
			giveRe: tpl.Replaces{
				Code: "  qwertyuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM_  ",
			},
			wantResult: []byte("\t  qwertyuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM_  \t"),
		},
	} {
		tt := tt

		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.wantResult, tpl.Replace(tt.giveIn, tt.giveRe))
		})
	}
}

func BenchmarkReplace(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		tpl.Replace([]byte("-- {{ code }} {{code}} __ {{message}} {{ description }} "), tpl.Replaces{
			Code:        "123",
			Message:     "message",
			Description: "desc",
		})
	}
}
