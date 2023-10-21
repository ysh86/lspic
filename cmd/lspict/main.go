package main

import (
	"fmt"
	"io"
	"os"

	"github.com/ysh86/lspic/pict"
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

	pictFile, err := pict.NewFile(io.NewSectionReader(file, 0, stat.Size()))
	if err != nil {
		panic(err)
	}
	if err := pictFile.Parse(); err != nil {
		panic(err)
	}

	// dump
	fmt.Printf("file: frame %+v\n", pictFile.Frame)
	for _, op := range pictFile.Ops {
		op.Dump()
	}
}
