package jpeg

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/ysh86/lspic/tiff"
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

var markerSegmentName map[uint16]string

func init() {
	markerSegmentName = map[uint16]string{
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

// Segment is a marker segment of jpeg.
type Segment struct {
	Marker uint16
	Length int64

	payloadFileOffset int64
	reader            *io.SectionReader

	parsedData Segmenter
}

// Parse parses a marker segment of jpeg.
func (s *Segment) Parse() error {
	switch s.Marker {
	case APP1:
		s.parsedData = &APP1Data{}
	case APP0:
		s.parsedData = &APP0Data{}
	default:
		s.parsedData = &SegmentData{}
	}

	return s.parsedData.Parse(s)
}

// Name generates the name string of the segment.
func (s *Segment) Name() string {
	name, ok := markerSegmentName[s.Marker]
	if !ok {
		name = fmt.Sprintf("%x", s.Marker)
	}
	return name
}

// String makes Segment satisfy the Stringer interface.
func (s *Segment) String() string {
	return fmt.Sprintf("%s: %08x, %d[bytes]", s.Name(), s.payloadFileOffset, s.Length)
}

// DumpTo prints the content of Segment.
func (s *Segment) DumpTo(w, wXmp io.Writer) {
	fmt.Fprintln(w, s)
	fmt.Fprint(w, s.parsedData)

	// XMP
	if d, ok := s.parsedData.(*APP1Data); ok {
		d.dumpXmpPacketTo(wXmp)
	}
}

// SplitTo writes raw data to w.
func (s *Segment) SplitTo(w io.Writer, offset, length int64) (int64, error) {
	if _, ok := s.parsedData.(*SegmentData); ok {
		s.reader.Seek(offset, io.SeekStart)
		return io.CopyN(w, s.reader, length)
	}
	return 0, fmt.Errorf("invalid segment")
}

// Segmenter is the interface of Segment parser
type Segmenter interface {
	Parse(segment *Segment) error
	fmt.Stringer
}

// APP1Data is the Application Segment 1 (Exif)
type APP1Data struct {
	identifier string

	// Exif
	Exif *tiff.File

	// XMP
	XmpPacket []byte
	// ExtendedXMP
	md5Digest         [32]byte
	fullLength        int64
	offsetThisPortion int64
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

// SegmentData is a dummy(unknown) segment
type SegmentData struct {
	// dummy
}

// Parse parses APP1 data.
func (d *APP1Data) Parse(segment *Segment) error {
	sr := segment.reader

	var ident [6]byte
	if _, err := sr.Read(ident[:]); err != nil {
		return err
	}
	if bytes.Equal(ident[:], []byte{'E', 'x', 'i', 'f', 0, 0}) {
		// Exif
		d.identifier = string(ident[0:4])

		// 0th IFD, 1st IFD (optional): thumbnail
		offset := int64(len(ident))
		length := segment.Length - int64(len(ident))
		var err error
		d.Exif, err = tiff.NewFile(io.NewSectionReader(sr, offset, length), segment.payloadFileOffset+offset)
		if err != nil {
			return err
		}
		return d.Exif.Parse()
	} else {
		// XMP
		longIdent := make([]byte, 0, 64)
		longIdent = append(longIdent, ident[:]...)
		for {
			var b byte
			err := binary.Read(sr, binary.BigEndian, &b)
			if err != nil || b == 0 {
				break
			}
			longIdent = append(longIdent, b)
		}

		d.identifier = string(longIdent)

		if d.identifier == "http://ns.adobe.com/xap/1.0/" {
			payload, err := io.ReadAll(sr)
			if err != nil {
				return err
			}
			d.XmpPacket = payload
			return nil
		}

		if d.identifier == "http://ns.adobe.com/xmp/extension/" {
			// GUID
			_, err := sr.Read(d.md5Digest[:])
			if err != nil {
				return err
			}

			// Full length
			var l uint32
			err = binary.Read(sr, binary.BigEndian, &l)
			if err != nil {
				return err
			}
			d.fullLength = int64(l)

			// Offset
			var o uint32
			err = binary.Read(sr, binary.BigEndian, &o)
			if err != nil {
				return err
			}
			d.offsetThisPortion = int64(o)

			payload, err := io.ReadAll(sr)
			if err != nil {
				return err
			}
			d.XmpPacket = payload
			return nil
		}

		return errors.New("invalid ident of APP1: " + d.identifier)
	}
}

// String makes APP1Data satisfy the Stringer interface.
func (d *APP1Data) String() string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("  identifier: %s\n", d.identifier))

	// XMP
	if len(d.XmpPacket) != 0 {
		if d.fullLength == 0 {
			buf.WriteString("  XMP packet: 1st\n")
		} else {
			buf.WriteString(fmt.Sprintf("  XMP packet: %s, %d/%d, %d[bytes]\n", string(d.md5Digest[:]), d.offsetThisPortion, d.fullLength, len(d.XmpPacket)))
		}
		return buf.String()
	}

	// Exif
	buf.WriteString(d.Exif.String())

	return buf.String()
}

func (d *APP1Data) dumpXmpPacketTo(w io.Writer) {
	if len(d.XmpPacket) > 0 {
		fmt.Fprint(w, string(d.XmpPacket))
	}
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
