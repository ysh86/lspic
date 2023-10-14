package tiff

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

type File struct {
	// Header
	byteOrder  binary.ByteOrder
	offsetNext int64

	IFDs [][]*IFDEntry

	reader       *io.SectionReader
	globalOffset int64
}

func NewFile(sr *io.SectionReader, globalOffset int64) (*File, error) {
	f := &File{reader: sr, globalOffset: globalOffset}
	return f, nil
}

func (f *File) Parse() error {
	err := f.parseFileHeader()
	if err != nil {
		return err
	}
	err = f.parseIFDs()
	if err != nil {
		return err
	}

	return nil
}

func (f *File) String() string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("  byte order: %s\n", f.byteOrder))
	for i, entries := range f.IFDs {
		buf.WriteString(fmt.Sprintf("    ========= IFD: %d\n", i))
		for _, entry := range entries {
			buf.WriteString(entry.String())
			buf.WriteString("    ----\n")
		}
	}
	return buf.String()
}

func (f *File) parseFileHeader() error {
	var endian uint16
	if err := binary.Read(f.reader, binary.BigEndian, &endian); err != nil {
		return err
	}
	var byteOrder binary.ByteOrder
	if endian == 0x4949 {
		byteOrder = binary.LittleEndian
	} else if endian == 0x4d4d {
		byteOrder = binary.BigEndian
	} else {
		return errors.New("invalid byte order")
	}

	// version
	var value42 uint16
	if err := binary.Read(f.reader, byteOrder, &value42); err != nil {
		return err
	}
	if value42 != 0x002a {
		return errors.New("invalid 42")
	}

	var offsetNext uint32
	if err := binary.Read(f.reader, byteOrder, &offsetNext); err != nil {
		return err
	}

	f.byteOrder = byteOrder
	f.offsetNext = int64(offsetNext)
	if _, err := f.reader.Seek(f.offsetNext, io.SeekStart); err != nil {
		return errors.New("invalid offset of 0th IFD")
	}
	f.IFDs = [][]*IFDEntry{}

	return nil
}

func (f *File) parseIFDs() error {
	for {
		offsetNext, entries, err := parseIFD(f.reader, f.byteOrder, f.globalOffset)
		if err != nil {
			return err
		}
		f.IFDs = append(f.IFDs, entries)

		if offsetNext == 0 {
			// 0 means the end of IFDs.
			break
		}
		if _, err := f.reader.Seek(int64(offsetNext), io.SeekStart); err != nil {
			return errors.New("invalid offset of next IFD")
		}
	}

	return nil
}
