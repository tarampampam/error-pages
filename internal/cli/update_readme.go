//go:build ignore
// +build ignore

package main

import (
	"os"

	"gh.tarampamp.am/error-pages/internal/cli"
)

const readmePath = "../../README.md"

func main() {
	if stat, err := os.Stat(readmePath); err == nil && stat.Mode().IsRegular() {
		var app = cli.NewApp("error-pages")

		if err = app.ToTabularToFileBetweenTags("error-pages", readmePath); err != nil {
			panic(err)
		}
	} else if err != nil {
		println("readme file not found, cli docs not updated:", err.Error())
	}
}
