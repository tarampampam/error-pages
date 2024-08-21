package serve_test

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"gh.tarampamp.am/error-pages/internal/cli/serve"
	"gh.tarampamp.am/error-pages/internal/logger"
)

func TestCommand_Run(t *testing.T) {
	t.Parallel()

	var (
		port = getFreeTcpPort(t)
		cmd  = serve.NewCommand(logger.NewNop())
	)

	var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var ch = make(chan error, 1)

	go func() {
		defer close(ch)

		ch <- cmd.Run(ctx, []string{
			"serve",
			"--port", strconv.Itoa(int(port)),
			"--add-template", "./testdata/foo-template.html",
			"--disable-template", "ghost",
			"--disable-template", "<unknown>",
			"--add-code", "200=Code/Description",
			"--json-format", "json format",
			"--xml-format", "xml format",
			"--plaintext-format", "plaintext format",
			"--template-name", "foo-template",
			"--disable-l10n",
			"--default-error-page", "503",
			"--send-same-http-code",
			"--show-details",
			"--proxy-headers", "X-Forwarded-For,X-Forwarded-Proto",
			"--rotation-mode", "random-on-each-request",
		})
	}()

	var connected bool

	for {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), time.Second)
		if err == nil {
			connected = true

			require.NoError(t, conn.Close())

			break
		} else {
			t.Log(err)
		}

		select {
		case <-ctx.Done():
			t.Fatal("timeout")
		case chErr := <-ch:
			require.NoError(t, chErr)
		case <-time.After(10 * time.Millisecond):
		}
	}

	require.True(t, connected, "server is not running")
}

// getFreeTcpPort is a helper function to get a free TCP port number.
func getFreeTcpPort(t *testing.T) uint16 {
	t.Helper()

	l, lErr := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, lErr)

	port := l.Addr().(*net.TCPAddr).Port
	require.NoError(t, l.Close())

	// make sure port is closed
	for {
		conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", port))
		if err != nil {
			break
		}

		require.NoError(t, conn.Close())
		<-time.After(5 * time.Millisecond)
	}

	return uint16(port) //nolint:gosec
}
