package pict

import (
	"encoding/binary"
	"errors"
	"image"
	"io"
)

type File struct {
	Frame image.Rectangle

	Ops []Operator

	reader *io.SectionReader
}

func NewFile(sr *io.SectionReader) (*File, error) {
	f := &File{reader: sr}
	return f, nil
}

func (f *File) Parse() error {
	// skip 512
	_, err := f.reader.Seek(0x200, io.SeekCurrent)
	if err != nil {
		return err
	}

	// length16
	var length16 uint16
	err = binary.Read(f.reader, binary.BigEndian, &length16)
	if err != nil {
		return err
	}

	// Frame
	var temp16 uint16
	if err := binary.Read(f.reader, binary.BigEndian, &temp16); err != nil {
		return err
	}
	f.Frame.Min.Y = int(temp16)
	if err := binary.Read(f.reader, binary.BigEndian, &temp16); err != nil {
		return err
	}
	f.Frame.Min.X = int(temp16)
	if err := binary.Read(f.reader, binary.BigEndian, &temp16); err != nil {
		return err
	}
	f.Frame.Max.Y = int(temp16)
	if err := binary.Read(f.reader, binary.BigEndian, &temp16); err != nil {
		return err
	}
	f.Frame.Max.X = int(temp16)

	// Ops v2
	for {
		var opcode uint16
		if err := binary.Read(f.reader, binary.BigEndian, &opcode); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		op, err := NewOp(opcode, f.reader)
		if err != nil {
			return err
		}
		f.Ops = append(f.Ops, op)
	}

	// check
	if length16 != uint16(f.reader.Size()-512) {
		return errors.New("size mismatch")
	}

	return nil
}
