package png

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

type Chunk struct {
	reader io.Reader
}

func (c *Chunk) Dump() {
	var length int32
	binary.Read(c.reader, binary.BigEndian, &length)

	chunkType := make([]byte, 4)
	c.reader.Read(chunkType)

	fmt.Printf("chunk '%v' (%d bytes)", string(chunkType), length)

	if bytes.Equal(chunkType, []byte("IHDR")) {
		if length == 13 {
			var v4 int32
			var v1 int8
			fmt.Printf(": ")
			binary.Read(c.reader, binary.BigEndian, &v4)
			fmt.Printf("Width = %d, ", v4)
			binary.Read(c.reader, binary.BigEndian, &v4)
			fmt.Printf("Height = %d, ", v4)
			binary.Read(c.reader, binary.BigEndian, &v1)
			fmt.Printf("Bit depth = %d, ", v1)
			binary.Read(c.reader, binary.BigEndian, &v1)
			fmt.Printf("Color type = %d, ", v1)
			binary.Read(c.reader, binary.BigEndian, &v1)
			fmt.Printf("Compression method = %d, ", v1)
			binary.Read(c.reader, binary.BigEndian, &v1)
			fmt.Printf("Filter method = %d, ", v1)
			binary.Read(c.reader, binary.BigEndian, &v1)
			fmt.Printf("Interlace method = %d\n", v1)
		} else {
			fmt.Printf(": corrupted!\n")
		}
	} else if bytes.Equal(chunkType, []byte("sRGB")) {
		if length == 1 {
			var v1 int8
			fmt.Printf(": ")
			binary.Read(c.reader, binary.BigEndian, &v1)
			fmt.Printf("Rendering intent = %d\n", v1)
		} else {
			fmt.Printf(": corrupted!\n")
		}
	} else if bytes.Equal(chunkType, []byte("tEXt")) {
		if length > 0 {
			rawText := make([]byte, length)
			c.reader.Read(rawText)
			fmt.Printf(": \"%s\"\n", string(rawText))
		} else {
			fmt.Printf(": corrupted!\n")
		}
	} else {
		fmt.Printf("\n")
	}

	// TODO: CRC
}
