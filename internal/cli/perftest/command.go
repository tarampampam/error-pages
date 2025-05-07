package perftest

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"time"

	"github.com/urfave/cli/v3"

	"gh.tarampamp.am/error-pages/internal/cli/shared"
)

const wrkOneCodeTestLua = `
local formats = { 'application/json', 'application/xml', 'text/html', 'text/plain' }

request = function()
		wrk.headers["User-Agent"] = "wrk"
    wrk.headers["X-Namespace"] = "NAMESPACE_" .. tostring(math.random(0, 99999999))
    wrk.headers["X-Request-ID"] = "REQ_ID_" .. tostring(math.random(0, 99999999))
    wrk.headers["Content-Type"] = formats[ math.random( 0, #formats - 1 ) ]

    return wrk.format("GET", "/500.html?rnd=" .. tostring(math.random(0, 99999999)), nil, nil)
end
`

//nolint:lll
const bombDifferentCodes = `
local formats = { 'application/json', 'application/xml', 'text/html', 'text/plain' }

request = function()
		wrk.headers["User-Agent"] = "wrk"
    wrk.headers["X-Namespace"] = "NAMESPACE_" .. tostring(math.random(0, 99999999))
    wrk.headers["X-Request-ID"] = "REQ_ID_" .. tostring(math.random(0, 99999999))
    wrk.headers["Content-Type"] = formats[ math.random( 0, #formats - 1 ) ]

    return wrk.format("GET", "/" .. tostring(math.random(400, 599)) .. ".html?rnd=" .. tostring(math.random(0, 99999999)), nil, nil)
end
`

// NewCommand creates `perftest` command.
func NewCommand() *cli.Command { //nolint:funlen
	var (
		portFlag     = shared.ListenPortFlag
		durationFlag = cli.DurationFlag{
			Name:    "duration",
			Aliases: []string{"d"},
			Usage:   "Duration of test",
			Value:   15 * time.Second, //nolint:mnd
			Validator: func(d time.Duration) error {
				if d <= time.Second {
					return errors.New("duration can't be less than 1 second")
				}

				return nil
			},
		}
		threadsFlag = cli.UintFlag{
			Name:    "threads",
			Aliases: []string{"t"},
			Usage:   "Number of threads to use",
			Value:   max(2, uint(math.Round(float64(runtime.NumCPU())/1.3))), //nolint:mnd
			Validator: func(u uint) error {
				if u == 0 {
					return errors.New("threads number can't be zero")
				} else if u > math.MaxUint16 {
					return errors.New("threads number can't be greater than 65535")
				}

				return nil
			},
		}
		connectionsFlag = cli.UintFlag{
			Name:    "connections",
			Aliases: []string{"c"},
			Usage:   "Number of connections to keep open",
			Value:   max(16, uint(runtime.NumCPU()*25)), //nolint:gosec,mnd
			Validator: func(u uint) error {
				if u == 0 {
					return errors.New("threads number can't be zero")
				} else if u > math.MaxUint16 {
					return errors.New("threads number can't be greater than 65535")
				}

				return nil
			},
		}
	)

	return &cli.Command{
		Name:    "perftest",
		Aliases: []string{"perf", "benchmark", "bench"},
		Hidden:  true,
		Usage:   "Performance (load) test for the HTTP server (locally installed wrk is required)",
		Action: func(ctx context.Context, c *cli.Command) error {
			var wrkBinPath, lErr = exec.LookPath("wrk")
			if lErr != nil {
				return fmt.Errorf("seems like wrk (https://github.com/wg/wrk) is not installed: %w", lErr)
			}

			var runTest = func(scriptContent string) error {
				if stdOut, stdErr, err := wrkRunTest(ctx,
					wrkBinPath,
					uint16(c.Uint(threadsFlag.Name)),     //nolint:gosec
					uint16(c.Uint(connectionsFlag.Name)), //nolint:gosec
					c.Duration(durationFlag.Name),
					uint16(c.Uint(portFlag.Name)), //nolint:gosec
					scriptContent,
				); err != nil {
					var errData, _ = io.ReadAll(stdErr)

					return fmt.Errorf("failed to execute the test: %w (%s)", err, string(errData))
				} else {
					var outData, _ = io.ReadAll(stdOut)

					printf("Test completed successfully. Here is the output:\n\n%s\n", string(outData))
				}

				return nil
			}

			printf("Starting the test to bomb ONE PAGE (code). Please, be patient...\n")

			if err := runTest(wrkOneCodeTestLua); err != nil {
				return err
			}

			printf("Starting the test to bomb DIFFERENT PAGES (codes). Please, be patient...\n")

			if err := runTest(bombDifferentCodes); err != nil {
				return err
			}

			return nil
		},
		Flags: []cli.Flag{
			&portFlag,
			&durationFlag,
			&threadsFlag,
			&connectionsFlag,
		},
	}
}

func printf(format string, args ...any) { fmt.Printf(format, args...) } //nolint:forbidigo

func wrkRunTest(
	ctx context.Context,
	wrkBinPath string,
	threadsCount, connectionsCount uint16,
	duration time.Duration,
	port uint16,
	scriptContent string,
) (io.Reader, io.Reader, error) {
	var tmpFile, tErr = os.CreateTemp("", "ep-perf-one-page")
	if tErr != nil {
		return nil, nil, fmt.Errorf("failed to create a temporary file: %w", tErr)
	}

	defer func() {
		_ = tmpFile.Close()
		_ = os.Remove(tmpFile.Name())
	}()

	if _, err := tmpFile.WriteString(scriptContent); err != nil {
		return nil, nil, fmt.Errorf("failed to write to a temporary file: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		return nil, nil, err
	}

	var stdout, stderr bytes.Buffer

	var cmd = exec.CommandContext(ctx, wrkBinPath, //nolint:gosec
		"--timeout", "1s",
		"--threads", strconv.FormatUint(uint64(threadsCount), 10),
		"--connections", strconv.FormatUint(uint64(connectionsCount), 10),
		"--duration", duration.String(),
		"--script", tmpFile.Name(),
		fmt.Sprintf("http://127.0.0.1:%d/", port),
	)

	cmd.Stdout, cmd.Stderr = &stdout, &stderr

	return &stdout, &stderr, cmd.Run() // execute
}
