package main

import (
	crypto "crypto/rand"
	"encoding/binary"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/fatih/color"
	"go.uber.org/automaxprocs/maxprocs"

	"github.com/tarampampam/error-pages/internal/cli"
)

// set GOMAXPROCS to match Linux container CPU quota.
var _, _ = maxprocs.Set(maxprocs.Min(1), maxprocs.Logger(func(_ string, _ ...any) {}))

// exitFn is a function for application exiting.
var exitFn = os.Exit //nolint:gochecknoglobals

// main CLI application entrypoint.
func main() { exitFn(run()) }

// run this CLI application.
// Exit codes documentation: <https://tldp.org/LDP/abs/html/exitcodes.html>
func run() int {
	var b [8]byte

	// seed random number generator
	if _, err := crypto.Read(b[:]); err == nil {
		rand.Seed(int64(binary.LittleEndian.Uint64(b[:]))) // https://stackoverflow.com/a/54491783/2252921
	} else {
		rand.Seed(time.Now().UnixNano()) // fallback
	}

	if err := (cli.NewApp(filepath.Base(os.Args[0]))).Run(os.Args); err != nil {
		_, _ = color.New(color.FgHiRed, color.Bold).Fprintln(os.Stderr, err.Error())

		return 1
	}

	return 0
}
