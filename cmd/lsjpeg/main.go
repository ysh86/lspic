package main

import (
	"fmt"
	"io"
	"os"
	"private/lsjpeg"
)

func main() {
	// args
	var (
		srcFile string
	)
	if len(os.Args) > 1 && os.Args[1] != "-h" {
		srcFile = os.Args[1]
	} else {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintln(os.Stderr, "  string")
		fmt.Fprintln(os.Stderr, "\tsrc file")
		return
	}

	file, err := os.Open(srcFile)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		panic(err)
	}

	jpegFile, err := lsjpeg.NewFile(io.NewSectionReader(file, 0, stat.Size()))
	if err != nil {
		panic(err)
	}
	if err := jpegFile.Parse(); err != nil {
		panic(err)
	}
	for _, s := range jpegFile.Segments {
		if err := s.Parse(); err != nil {
			panic(err)
		}
		s.DumpTo(os.Stderr)
	}
}
