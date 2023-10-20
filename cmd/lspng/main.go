package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
)

func dumpChunk(chunk io.Reader) {
	var length int32
	binary.Read(chunk, binary.BigEndian, &length)

	chunkType := make([]byte, 4)
	chunk.Read(chunkType)

	fmt.Printf("chunk '%v' (%d bytes)", string(chunkType), length)

	if bytes.Equal(chunkType, []byte("IHDR")) {
		if length == 13 {
			var v4 int32
			var v1 int8
			fmt.Printf(": ")
			binary.Read(chunk, binary.BigEndian, &v4)
			fmt.Printf("Width = %d, ", v4)
			binary.Read(chunk, binary.BigEndian, &v4)
			fmt.Printf("Height = %d, ", v4)
			binary.Read(chunk, binary.BigEndian, &v1)
			fmt.Printf("Bit depth = %d, ", v1)
			binary.Read(chunk, binary.BigEndian, &v1)
			fmt.Printf("Color type = %d, ", v1)
			binary.Read(chunk, binary.BigEndian, &v1)
			fmt.Printf("Compression method = %d, ", v1)
			binary.Read(chunk, binary.BigEndian, &v1)
			fmt.Printf("Filter method = %d, ", v1)
			binary.Read(chunk, binary.BigEndian, &v1)
			fmt.Printf("Interlace method = %d\n", v1)
		} else {
			fmt.Printf(": corrupted!\n")
		}
	} else if bytes.Equal(chunkType, []byte("sRGB")) {
		if length == 1 {
			var v1 int8
			fmt.Printf(": ")
			binary.Read(chunk, binary.BigEndian, &v1)
			fmt.Printf("Rendering intent = %d\n", v1)
		} else {
			fmt.Printf(": corrupted!\n")
		}
	} else if bytes.Equal(chunkType, []byte("tEXt")) {
		if length > 0 {
			rawText := make([]byte, length)
			chunk.Read(rawText)
			fmt.Printf(": \"%s\"\n", string(rawText))
		} else {
			fmt.Printf(": corrupted!\n")
		}
	} else {
		fmt.Printf("\n")
	}

	// TODO: CRC
}

func parseChunks(sr *io.SectionReader) (chunks []io.Reader, err error) {
	signature := make([]byte, 8)
	n, err := sr.Read(signature)
	if err != nil || !bytes.Equal(signature, []byte{137, 80, 78, 71, 13, 10, 26, 10}) {
		return nil, errors.New("invalid signature")
	}

	offset := int64(n)
	for {
		var length int32
		err = binary.Read(sr, binary.BigEndian, &length)
		if err != nil {
			break
		}
		// chunk = length, type, data, CRC
		chunks = append(chunks, io.NewSectionReader(sr, offset, 4+4+int64(length)+4))
		offset, err = sr.Seek(4+int64(length)+4, io.SeekCurrent)
		if err != nil {
			break
		}
	}

	return chunks, err
}

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

	chunks, err := parseChunks(io.NewSectionReader(file, 0, stat.Size()))
	if err != io.EOF {
		panic(err)
	}
	for _, chunk := range chunks {
		dumpChunk(chunk)
	}
}
