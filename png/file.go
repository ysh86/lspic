package png

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
)

type File struct {
	Chunks []*io.SectionReader

	reader *io.SectionReader
}

func NewFile(sr *io.SectionReader) (*File, error) {
	f := &File{reader: sr}
	return f, nil
}

func (f *File) Parse() error {
	signature := make([]byte, 8)
	n, err := f.reader.Read(signature)
	if err != nil || !bytes.Equal(signature, []byte{137, 80, 78, 71, 13, 10, 26, 10}) {
		return errors.New("invalid signature")
	}

	offset := int64(n)
	for {
		var length int32
		err = binary.Read(f.reader, binary.BigEndian, &length)
		if err != nil {
			break
		}
		// chunk = length, type, data, CRC
		f.Chunks = append(f.Chunks, io.NewSectionReader(f.reader, offset, 4+4+int64(length)+4))
		offset, err = f.reader.Seek(4+int64(length)+4, io.SeekCurrent)
		if err != nil {
			break
		}
	}

	if err == io.EOF {
		return nil
	}
	return err
}
