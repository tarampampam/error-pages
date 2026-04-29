//go:build readme

package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"gh.tarampamp.am/error-pages/v4/cmd/error-pages/app"
)

func main() {
	var outFile string

	flag.StringVar(&outFile, "out", "./../README.md", "output file path")
	flag.Parse()

	if stat, statErr := os.Stat(outFile); statErr == nil && stat.Mode().IsRegular() {
		var help = app.NewApp("error-pages").Help()

		if err := replaceWith(outFile, help); err != nil {
			panic(err)
		}
	} else if statErr != nil {
		fmt.Println("⚠ readme file not found, cli docs not updated:", statErr.Error())
	}
}

func replaceWith(filePath string, content string) error {
	const start, end = "<!--GENERATED:SERVER_CLI-->", "<!--/GENERATED:SERVER_CLI-->"

	// read original file content
	original, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	from, to := strings.Index(string(original), start), strings.Index(string(original), end)
	if from == -1 || to == -1 {
		return errors.New("start or end tag not found")
	}

	// write updated content to file
	if err = os.WriteFile(filePath, []byte(strings.Join([]string{
		string(original[:from+len(start)]),
		"```", content, "```",
		string(original[to:]),
	}, "\n")), 0o664); err != nil {
		return err
	}

	return nil
}
