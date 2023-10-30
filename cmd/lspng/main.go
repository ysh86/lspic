package main

import (
	"fmt"
	"io"
	"os"

	"github.com/ysh86/lspic/png"
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

	pngFile, err := png.NewFile(io.NewSectionReader(file, 0, stat.Size()))
	if err != nil {
		panic(err)
	}
	if err := pngFile.Parse(); err != nil {
		panic(err)
	}

	// dump
	for _, chunk := range pngFile.Chunks {
		chunk.Dump()
	}
}
