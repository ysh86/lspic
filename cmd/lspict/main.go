package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

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
	hasQT := false
	fmt.Printf("file: frame %+v\n", pictFile.Frame)
	for _, op := range pictFile.Ops {
		op.Dump()
		if _, ok := op.(*pict.OpQTcomp); ok {
			hasQT = true
		}
	}
	if !hasQT {
		return
	}

	// dump QT
	fmt.Printf("dump QT:\n")
	n := 0
	for _, op := range pictFile.Ops {
		if qt, ok := op.(*pict.OpQTcomp); ok {
			dir, file := filepath.Split(srcFile)
			ext := filepath.Ext(file)
			file, _ = strings.CutSuffix(file, ext)
			ext = fmt.Sprintf("_%d.jpg", n)
			n += 1
			full := filepath.Join(dir, file+ext)
			w, err := os.Create(full)
			if err != nil {
				panic(err)
			}
			if _, err := qt.DumpTo(w); err != nil {
				panic(err)
			}
			w.Close()
			fmt.Printf("  %s\n", full)
		}
	}
}
