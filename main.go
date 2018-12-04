package main

import (
	"fmt"
	"io"
	"os"
	"private/lsjpeg/jpeg"
)

func dumpSegment(segment *jpeg.Segment) {
	fmt.Println(segment)

	var data jpeg.Segmenter
	switch segment.Marker {
	case jpeg.APP1:
		data = &jpeg.APP1Data{}
	case jpeg.APP0:
		data = &jpeg.APP0Data{}
	default:
		data = &jpeg.SegmentData{}
	}
	err := data.Parse(segment)
	if err != nil {
		fmt.Printf("  %v\n", err)
	} else {
		fmt.Print(data)
	}
}

func main() {
	// args
	var (
		srcFile string
	)
	if len(os.Args) > 1 && os.Args[1] != "-h" {
		srcFile = os.Args[1]
	} else {
		fmt.Printf("Usage of %s:\n", os.Args[0])
		fmt.Println("  string")
		fmt.Println("\tsrc file")
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

	jpegFile, err := jpeg.NewFile(io.NewSectionReader(file, 0, stat.Size()))
	if err != nil {
		panic(err)
	}
	if err := jpegFile.Parse(); err != nil {
		panic(err)
	}
	for _, s := range jpegFile.Segments {
		dumpSegment(s)
	}
}
