package perftest

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/urfave/cli/v3"

	"gh.tarampamp.am/error-pages/internal/cli/shared"
	"gh.tarampamp.am/error-pages/internal/logger"
)

// NewCommand creates `perftest` command.
func NewCommand(log *logger.Logger) *cli.Command { //nolint:funlen,gocognit
	var (
		portFlag     = shared.ListenPortFlag
		durationFlag = cli.DurationFlag{
			Name:    "duration",
			Aliases: []string{"d"},
			Usage:   "duration of the test",
			Value:   10 * time.Second, //nolint:mnd
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
			Usage:   "number of threads",
			Value:   max(2, uint64(runtime.NumCPU()/2)), //nolint:mnd
			Validator: func(u uint64) error {
				if u == 0 {
					return errors.New("threads number can't be zero")
				}

				return nil
			},
		}
	)

	return &cli.Command{
		Name:    "perftest",
		Aliases: []string{"perf", "test"},
		Hidden:  true,
		Usage:   "Simple performance (load) test for the HTTP server",
		Action: func(ctx context.Context, c *cli.Command) error {
			var (
				perfCtx, cancel = context.WithTimeout(ctx, c.Duration(durationFlag.Name))
				startedAt       = time.Now()

				wg      sync.WaitGroup
				success atomic.Uint64
				failed  atomic.Uint64
			)

			defer func() {
				cancel()

				log.Info("Summary",
					logger.Uint64("success", success.Load()),
					logger.Uint64("failed", failed.Load()),
					logger.Duration("duration", time.Since(startedAt)),
					logger.Float64("RPS", float64(success.Load()+failed.Load())/time.Since(startedAt).Seconds()),
					logger.Float64("errors rate", float64(failed.Load())/float64(success.Load()+failed.Load())*100),
				)
			}()

			log.Info("Running test",
				logger.Uint64("threads", c.Uint(threadsFlag.Name)),
				logger.Duration("duration", c.Duration(durationFlag.Name)),
			)

			var httpClient = &http.Client{
				Transport: &http.Transport{MaxConnsPerHost: max(2, int(c.Uint(threadsFlag.Name))-1)}, //nolint:mnd
				Timeout:   c.Duration(durationFlag.Name) + time.Second,
			}

			for i := uint64(0); i < c.Uint(threadsFlag.Name); i++ {
				wg.Add(1)

				go func(log *logger.Logger) {
					defer wg.Done()

					if perfCtx.Err() != nil {
						return
					}

					var req, rErr = makeRequest(perfCtx, uint16(c.Uint(portFlag.Name)))
					if rErr != nil {
						log.Error("failed to create a new request", logger.Error(rErr))

						return
					}

					for {
						var sentAt = time.Now()

						var resp, respErr = httpClient.Do(req)
						if resp != nil {
							if _, err := io.Copy(io.Discard, resp.Body); err != nil && !errIsDone(err) {
								log.Error("failed to read response body", logger.Error(err))
							}

							if err := resp.Body.Close(); err != nil && !errIsDone(err) {
								log.Error("failed to close response body", logger.Error(err))
							}
						}

						if respErr != nil {
							if errIsDone(respErr) {
								return
							}

							log.Error("request failed", logger.Error(respErr))
							failed.Add(1)

							continue
						}

						log.Debug("Response received",
							logger.String("status", resp.Status),
							logger.Duration("duration", time.Since(sentAt)),
							logger.Int64("size", resp.ContentLength),
							logger.Uint64("success", success.Load()),
							logger.Uint64("failed", failed.Load()),
						)

						success.Add(1)
					}
				}(log.Named(fmt.Sprintf("thread-%d", i)))
			}

			wg.Wait()

			return nil
		},
		Flags: []cli.Flag{
			&portFlag,
			&durationFlag,
			&threadsFlag,
		},
	}
}

// randomIntBetween returns a random integer between min and max.
func randomIntBetween(min, max int) int { return min + rand.Intn(max-min) } //nolint:gosec

// makeRequest creates a new HTTP request for the performance test.
func makeRequest(ctx context.Context, port uint16) (*http.Request, error) {
	var req, rErr = http.NewRequestWithContext(ctx,
		http.MethodGet,
		fmt.Sprintf(
			"http://127.0.0.1:%d/%d.html?rnd=%d", // for load testing purposes only
			port,
			randomIntBetween(400, 418),       //nolint:mnd
			randomIntBetween(1, 999_999_999), //nolint:mnd
		),
		http.NoBody,
	)

	if rErr != nil {
		return nil, rErr
	}

	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("User-Agent", "perftest")
	req.Header.Set("X-Namespace", fmt.Sprintf("namespace-%d", randomIntBetween(1, 999_999_999))) //nolint:mnd
	req.Header.Set("X-Request-ID", fmt.Sprintf("req-id-%d", randomIntBetween(1, 999_999_999)))   //nolint:mnd

	var contentType string

	switch randomIntBetween(1, 4) { //nolint:mnd
	case 1:
		contentType = "application/json"
	case 2: //nolint:mnd
		contentType = "application/xml"
	case 3: //nolint:mnd
		contentType = "text/html"
	default:
		contentType = "text/plain"
	}

	req.Header.Set("Content-Type", contentType)

	return req, nil
}

// errIsDone checks if the error is a context.DeadlineExceeded or context.Canceled.
func errIsDone(err error) bool {
	return errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled)
}
