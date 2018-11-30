package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
)

// Marker Segment code
const (
	Unknown uint16 = 0

	SOI  uint16 = 0xffd8 // Start of Image
	APP0 uint16 = 0xffe0 // Application Segment 0 (JFIF)
	APP1 uint16 = 0xffe1 // Application Segment 1 (Exif)
	APP2 uint16 = 0xffe2 // Application Segment 2 (Flashpix)
	DQT  uint16 = 0xffdb // Define Quantization Table
	DHT  uint16 = 0xffc4 // Define Huffman Table
	DRI  uint16 = 0xffdd // Define Restart Interval
	SOF  uint16 = 0xffc0 // Start of Frame (Baseline DCT)
	SOS  uint16 = 0xffda // Start of Scan
	Data uint16 = 1
	EOI  uint16 = 0xffd9 // End of Image
)

// MarkerSegmentName is a map of Marker Segment name
var MarkerSegmentName map[uint16]string

func init() {
	MarkerSegmentName = map[uint16]string{
		Unknown: "Unknown",

		SOI:  "SOI",
		APP0: "APP0",
		APP1: "APP1",
		APP2: "APP2",
		DQT:  "DQT",
		DHT:  "DHT",
		DRI:  "DRI",
		SOF:  "SOF",
		SOS:  "SOS",
		Data: "Data",
		EOI:  "EOI",
	}
}

// Segment is a marker segment of jpeg
type Segment struct {
	marker uint16
	length int64
	reader io.ReadSeeker
}

// Segmenter is the interface of Segment parser
type Segmenter interface {
	Parse(segment *Segment) error
	fmt.Stringer
}

// APP1Data is the Application Segment 1 (Exif)
type APP1Data struct {
	identifier string
	byteOrder  binary.ByteOrder
}

// APP0Data is the Application Segment 0 (JFIF)
type APP0Data struct {
	identifier string
	version    uint16
	units      uint8
	xDensity   uint16
	yDensity   uint16
	xThumbnail uint8
	yThumbnail uint8
}

// SegmentData is a dummy segment
type SegmentData struct {
	// dummy
}

// IFD Tag
const (
	InvalidTag uint16 = 0

	ImageWidth  uint16 = 0x0100
	ImageLength uint16 = 0x0101
)

// IFD Type
const (
	InvalidType uint16 = iota

	BYTE      // []uint8
	ASCII     // []byte (NUL terminated)
	SHORT     // []uint16
	LONG      // []uint32
	RATIONAL  // []*big.Rat {num: uint32, den: uint32}
	SBYTE     // []int8
	UNDEFINED // []byte
	SSHORT    // []int16
	SLONG     // []int32
	SRATIONAL // []*big.Rat {num: int32, den: int32}
	// FLOAT
	// DOUBLE
)

// IFDEntry is the IFD entry
type IFDEntry struct {
	tag     uint16
	ifdType uint16
	count   uint32
	offset  uint32

	values []interface{}

	// cache
	elmSize int64
}

func (e *IFDEntry) elementSize() int64 {
	if e.elmSize != 0 {
		return e.elmSize
	}

	switch e.ifdType {
	case BYTE:
		e.elmSize = 1
	case ASCII:
		e.elmSize = 1
	case SHORT:
		e.elmSize = 2
	case LONG:
		e.elmSize = 4
	case RATIONAL:
		e.elmSize = 4 + 4
	case SBYTE:
		e.elmSize = 1
	case UNDEFINED:
		e.elmSize = 1
	case SSHORT:
		e.elmSize = 2
	case SLONG:
		e.elmSize = 4
	case SRATIONAL:
		e.elmSize = 4 + 4
	default:
		e.elmSize = 0
	}

	return e.elmSize
}

func (e *IFDEntry) parseValue4bytes(rs io.ReadSeeker, byteOrder binary.ByteOrder) error {
	pos, err := rs.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}

	e.values = make([]interface{}, 0, e.count)

	switch e.ifdType {
	case BYTE:
		var value uint8
		for i := e.count; i > 0; i-- {
			if err := binary.Read(rs, byteOrder, &value); err != nil {
				return err
			}
			e.values = append(e.values, value)
		}
	case ASCII:
		var value byte
		for i := e.count; i > 0; i-- {
			if err := binary.Read(rs, byteOrder, &value); err != nil {
				return err
			}
			e.values = append(e.values, value)
		}
	case SHORT:
		var value uint16
		for i := e.count; i > 0; i-- {
			if err := binary.Read(rs, byteOrder, &value); err != nil {
				return err
			}
			e.values = append(e.values, value)
		}
	case LONG:
		var value uint32
		for i := e.count; i > 0; i-- {
			if err := binary.Read(rs, byteOrder, &value); err != nil {
				return err
			}
			e.values = append(e.values, value)
		}
	case RATIONAL:
		// over 4bytes
		e.values = nil
	case SBYTE:
		var value int8
		for i := e.count; i > 0; i-- {
			if err := binary.Read(rs, byteOrder, &value); err != nil {
				return err
			}
			e.values = append(e.values, value)
		}
	case UNDEFINED:
		var value byte
		for i := e.count; i > 0; i-- {
			if err := binary.Read(rs, byteOrder, &value); err != nil {
				return err
			}
			e.values = append(e.values, value)
		}
	case SSHORT:
		var value int16
		for i := e.count; i > 0; i-- {
			if err := binary.Read(rs, byteOrder, &value); err != nil {
				return err
			}
			e.values = append(e.values, value)
		}
	case SLONG:
		var value int32
		for i := e.count; i > 0; i-- {
			if err := binary.Read(rs, byteOrder, &value); err != nil {
				return err
			}
			e.values = append(e.values, value)
		}
	case SRATIONAL:
		// over 4bytes
		e.values = nil
	default:
		e.values = nil
	}

	_, err = rs.Seek(pos+4, io.SeekStart)
	return err

	//return nil
}

// String makes IFDEntry satisfy the Stringer interface.
func (e *IFDEntry) String() string {
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("    Tag: %xh\n", e.tag))
	buf.WriteString(fmt.Sprintf("    Type: %d\n", e.ifdType))
	buf.WriteString(fmt.Sprintf("    Count: %d\n", e.count))
	buf.WriteString(fmt.Sprintf("    Offset: 0x%08x\n", e.offset))
	buf.WriteString(fmt.Sprintf("    Value: %+v\n", e.values))
	buf.WriteString(fmt.Sprintf("    ----\n"))

	return buf.String()
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

func parseJpegFile(sr *io.SectionReader) (segments []*Segment, e error) {
	var offset int64

	// SOI
	{
		marker, length, err := readMarkerLength(sr)
		if err != nil || marker != SOI || length != 0 {
			return segments, errors.New("expected SOI")
		}

		offset += 2 // 'marker uint16'
		segments = append(segments, &Segment{marker, int64(length), io.NewSectionReader(sr, offset, int64(length))})
	}

	// APP1(Exif) or APP0(JFIF)
	{
		marker, length, err := readMarkerLength(sr)
		if err != nil || (marker != APP1 && marker != APP0) || length < 2 {
			return segments, errors.New("expected APP1/0")
		}

		length -= 2 // length includes 'length uint16' itself.
		offset += 4 // 'marker uint16' + 'length uint16'
		segments = append(segments, &Segment{marker, int64(length), io.NewSectionReader(sr, offset, int64(length))})

		offset, err = sr.Seek(int64(length), io.SeekCurrent)
		if err != nil {
			return segments, errors.New("invalid length of APP1/0")
		}
	}

	// other segments
	for {
		marker, length, err := readMarkerLength(sr)
		if err != nil || length < 2 {
			return segments, errors.New("invalid segment")
		}

		length -= 2 // length includes 'length uint16' itself.
		offset += 4 // 'marker uint16' + 'length uint16'
		segments = append(segments, &Segment{marker, int64(length), io.NewSectionReader(sr, offset, int64(length))})

		offset, err = sr.Seek(int64(length), io.SeekCurrent)
		if err != nil {
			return segments, errors.New("invalid length of segment")
		}

		// SOS
		if marker == SOS {
			break
		}
	}

	// data
	{
		end, err := sr.Seek(-2, io.SeekEnd)
		length := end - offset
		if err != nil || length <= 0 {
			return segments, errors.New("invalid length of data")
		}

		segments = append(segments, &Segment{Data, length, io.NewSectionReader(sr, offset, length)})

		offset += length
	}

	// EOI
	{
		marker, length, err := readMarkerLength(sr)
		if err != nil || marker != EOI || length != 0 {
			return segments, errors.New("expected EOI")
		}

		offset += 2 // 'marker uint16'
		segments = append(segments, &Segment{marker, int64(length), io.NewSectionReader(sr, offset, int64(length))})
	}

	return segments, nil
}

// Parse parses APP1 data.
func (d *APP1Data) Parse(segment *Segment) error {
	rs := segment.reader

	var ident [6]byte
	if _, err := rs.Read(ident[:]); err != nil {
		return err
	}
	if !bytes.Equal(ident[:], []byte{'E', 'x', 'i', 'f', 0, 0}) {
		return errors.New("invalid ident of APP1")
	}
	d.identifier = string(ident[0:4])

	// TIFF File Header
	offsetHeader, err := rs.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}
	var endian uint16
	if err := binary.Read(rs, binary.BigEndian, &endian); err != nil {
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
	d.byteOrder = byteOrder

	var value42 uint16
	if err := binary.Read(rs, byteOrder, &value42); err != nil {
		return err
	}
	if value42 != 0x002a {
		return errors.New("invalid 42")
	}

	var offset uint32
	if err := binary.Read(rs, byteOrder, &offset); err != nil {
		return err
	}
	if _, err := rs.Seek(offsetHeader+int64(offset), io.SeekStart); err != nil {
		return errors.New("invalid offset of 1st IFD")
	}

	// 0th IFD, 1st IFD (optional): thumbnail
	for {
		fmt.Println("    ================")

		var num uint16
		if err := binary.Read(rs, byteOrder, &num); err != nil {
			return err
		}
		entries := make([]IFDEntry, num)
		for _, entry := range entries {
			// clear chache
			entry.elmSize = 0

			if err := binary.Read(rs, byteOrder, &entry.tag); err != nil {
				return err
			}
			if err := binary.Read(rs, byteOrder, &entry.ifdType); err != nil {
				return err
			}
			if err := binary.Read(rs, byteOrder, &entry.count); err != nil {
				return err
			}
			// Offset or Value
			totalBytes := entry.elementSize() * int64(entry.count)
			if totalBytes > 4 {
				// Offset
				if err := binary.Read(rs, byteOrder, &entry.offset); err != nil {
					return err
				}
				entry.offset += uint32(offsetHeader)
				entry.values = nil
			} else {
				// Value
				entry.offset = 0
				if err := entry.parseValue4bytes(rs, byteOrder); err != nil {
					return err
				}
			}

			fmt.Print(&entry)
		}

		var offset uint32
		if err := binary.Read(rs, byteOrder, &offset); err != nil {
			return err
		}
		if offset == 0 {
			break
		}
		if _, err := rs.Seek(offsetHeader+int64(offset), io.SeekStart); err != nil {
			return errors.New("invalid offset of next IFD")
		}
	}

	return nil
}

// String makes APP1Data satisfy the Stringer interface.
func (d *APP1Data) String() string {
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("  identifier: %s\n", d.identifier))
	buf.WriteString(fmt.Sprintf("  byte order: %s\n", d.byteOrder))

	return buf.String()
}

// Parse parses APP0 data.
func (d *APP0Data) Parse(segment *Segment) error {
	r := segment.reader

	var ident [5]byte
	if _, err := r.Read(ident[:]); err != nil {
		return err
	}
	if !bytes.Equal(ident[:], []byte{'J', 'F', 'I', 'F', 0}) {
		return errors.New("invalid ident of APP0")
	}
	d.identifier = string(ident[0:4])

	if err := binary.Read(r, binary.BigEndian, &d.version); err != nil {
		return err
	}
	if err := binary.Read(r, binary.BigEndian, &d.units); err != nil {
		return err
	}

	if err := binary.Read(r, binary.BigEndian, &d.xDensity); err != nil {
		return err
	}
	if err := binary.Read(r, binary.BigEndian, &d.yDensity); err != nil {
		return err
	}

	if err := binary.Read(r, binary.BigEndian, &d.xThumbnail); err != nil {
		return err
	}
	if err := binary.Read(r, binary.BigEndian, &d.yThumbnail); err != nil {
		return err
	}

	// TODO: Thumbnail (RGB xN)

	return nil
}

// String makes APP0Data satisfy the Stringer interface.
func (d *APP0Data) String() string {
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("  identifier: %s\n", d.identifier))
	buf.WriteString(fmt.Sprintf("  version: %04x\n", d.version))
	buf.WriteString(fmt.Sprintf("  units: %d\n", d.units))
	buf.WriteString(fmt.Sprintf("  WxH: %dx%d\n", d.xDensity, d.yDensity))
	buf.WriteString(fmt.Sprintf("  Thumbnail WxH: %dx%d\n", d.xThumbnail, d.yThumbnail))

	return buf.String()
}

// Parse is a dummy parser for generic segments.
func (d *SegmentData) Parse(segment *Segment) error {
	return nil
}

// String makes SegmentData satisfy the Stringer interface.
func (d *SegmentData) String() string {
	return ""
}

func dumpSegment(segment *Segment) {
	name, ok := MarkerSegmentName[segment.marker]
	if !ok {
		name = fmt.Sprintf("%x", segment.marker)
	}
	fmt.Printf("%s: %d[bytes]\n", name, segment.length)

	var data Segmenter
	switch segment.marker {
	case APP1:
		data = &APP1Data{}
	case APP0:
		data = &APP0Data{}
	default:
		data = &SegmentData{}
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

	segments, err := parseJpegFile(io.NewSectionReader(file, 0, stat.Size()))
	if err != nil {
		panic(err)
	}
	for _, s := range segments {
		dumpSegment(s)
	}
}
