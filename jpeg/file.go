package jpeg

import (
	"encoding/binary"
	"errors"
	"io"
)

// File is a struct for the JPEG file(JFIF).
type File struct {
	Segments []*Segment

	reader *io.SectionReader
}

// NewFile creates a new JPEG file struct.
func NewFile(sr *io.SectionReader) (*File, error) {
	f := &File{reader: sr}
	return f, nil
}

func readMarkerLength(r io.Reader) (marker uint16, length uint16, e error) {
	var buf uint16
	if err := binary.Read(r, binary.BigEndian, &buf); err != nil {
		return marker, length, err
	}

	marker = buf
	if marker != SOI && marker != EOI {
		if err := binary.Read(r, binary.BigEndian, &buf); err != nil {
			return marker, length, err
		}
		length = buf
	}

	return marker, length, nil
}

// Parse parses a JPEG file.
func (f *File) Parse() error {
	var offset int64

	// SOI
	{
		marker, length, err := readMarkerLength(f.reader)
		if err != nil || marker != SOI || length != 0 {
			return errors.New("expected SOI")
		}

		offset += 2 // 'marker uint16'
		seg := &Segment{marker, int64(length), offset, io.NewSectionReader(f.reader, offset, int64(length)), nil}
		if err := seg.Parse(); err != nil {
			return err
		}
		f.Segments = append(f.Segments, seg)
	}

	// APP1(Exif) or APP0(JFIF)
	{
		marker, length, err := readMarkerLength(f.reader)
		if err != nil || (marker != APP1 && marker != APP0) || length < 2 {
			return errors.New("expected APP1/0")
		}

		length -= 2 // length includes 'length uint16' itself.
		offset += 4 // 'marker uint16' + 'length uint16'
		seg := &Segment{marker, int64(length), offset, io.NewSectionReader(f.reader, offset, int64(length)), nil}
		if err := seg.Parse(); err != nil {
			return err
		}
		f.Segments = append(f.Segments, seg)

		offset, err = f.reader.Seek(int64(length), io.SeekCurrent)
		if err != nil {
			return errors.New("invalid length of APP1/0")
		}
	}

	// other segments
	for {
		marker, length, err := readMarkerLength(f.reader)
		if err != nil || length < 2 {
			return errors.New("invalid segment")
		}

		length -= 2 // length includes 'length uint16' itself.
		offset += 4 // 'marker uint16' + 'length uint16'
		seg := &Segment{marker, int64(length), offset, io.NewSectionReader(f.reader, offset, int64(length)), nil}
		if err := seg.Parse(); err != nil {
			return err
		}
		f.Segments = append(f.Segments, seg)

		offset, err = f.reader.Seek(int64(length), io.SeekCurrent)
		if err != nil {
			return errors.New("invalid length of segment")
		}

		// SOS
		if marker == SOS {
			break
		}
	}

	// data
	{
		end, err := f.reader.Seek(-2, io.SeekEnd)
		length := end - offset
		if err != nil || length <= 0 {
			return errors.New("invalid length of data")
		}

		seg := &Segment{Data, length, offset, io.NewSectionReader(f.reader, offset, length), nil}
		if err := seg.Parse(); err != nil {
			return err
		}
		f.Segments = append(f.Segments, seg)

		offset += length
	}

	// EOI
	{
		marker, length, err := readMarkerLength(f.reader)
		if err != nil || marker != EOI || length != 0 {
			return errors.New("expected EOI")
		}

		offset += 2 // 'marker uint16'
		seg := &Segment{marker, int64(length), offset, io.NewSectionReader(f.reader, offset, int64(length)), nil}
		if err := seg.Parse(); err != nil {
			return err
		}
		f.Segments = append(f.Segments, seg)
	}

	return nil
}
