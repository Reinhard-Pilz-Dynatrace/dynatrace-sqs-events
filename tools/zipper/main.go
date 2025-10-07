//go:build ignore

// tools/zipper/main.go
package main

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"time"
)

func main() {
	const out = "function.zip"

	args := os.Args[1:]
	if len(args) == 0 {
		args = []string{"bootstrap"}
	}

	outFile, err := os.Create(out)
	check(err)
	defer outFile.Close()

	zw := zip.NewWriter(outFile)
	defer zw.Close()

	constMode := os.FileMode(0o755)
	constTime := time.Unix(0, 0).UTC()

	for _, p := range args {
		addExecutable(zw, p, constMode, constTime)
	}
}

func addExecutable(zw *zip.Writer, path string, mode os.FileMode, mtime time.Time) {
	info, err := os.Stat(path)
	check(err)
	if info.IsDir() {
		return
	}

	fh, err := zip.FileInfoHeader(info)
	check(err)
	fh.Name = filepath.Base(path)
	fh.SetMode(mode)
	fh.Modified = mtime

	w, err := zw.CreateHeader(fh)
	check(err)

	in, err := os.Open(path)
	check(err)
	defer in.Close()

	_, err = io.Copy(w, in)
	check(err)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
