package pict

import (
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"io"
)

// opcode
const (
	Version      uint16 = 0x0011
	Header       uint16 = 0x0c00
	Clip         uint16 = 0x0001
	QTcomp       uint16 = 0x8200
	PackBitsRect uint16 = 0x0098
	EndPic       uint16 = 0x00ff
)

type Operator interface {
	Parse(r *io.SectionReader) error
	Dump()
}

type OpVersion struct {
	Opcode  uint16
	Version byte
}

func (o *OpVersion) Parse(r *io.SectionReader) error {
	var v uint16
	if err := binary.Read(r, binary.BigEndian, &v); err != nil {
		return err
	}
	if v != 0x02ff {
		return fmt.Errorf("invalid version: %04x", v)
	}

	o.Version = 2

	return nil
}
func (o *OpVersion) Dump() {
	fmt.Printf("  Op Version: %d\n", o.Version)
}

type OpHeader struct {
	Opcode  uint16
	ResH    uint32
	ResV    uint32
	SrcRect image.Rectangle
}

func (o *OpHeader) Parse(r *io.SectionReader) error {
	var temp16 uint16

	// version
	if err := binary.Read(r, binary.BigEndian, &temp16); err != nil {
		return err
	}
	if temp16 != 0xfffe {
		return fmt.Errorf("invalid version: %04x", temp16)
	}

	// reserved
	if err := binary.Read(r, binary.BigEndian, &temp16); err != nil {
		return err
	}

	// Res HxV
	if err := binary.Read(r, binary.BigEndian, &o.ResH); err != nil {
		return err
	}
	if err := binary.Read(r, binary.BigEndian, &o.ResV); err != nil {
		return err
	}

	// Src rect
	if err := binary.Read(r, binary.BigEndian, &temp16); err != nil {
		return err
	}
	o.SrcRect.Min.Y = int(temp16)
	if err := binary.Read(r, binary.BigEndian, &temp16); err != nil {
		return err
	}
	o.SrcRect.Min.X = int(temp16)
	if err := binary.Read(r, binary.BigEndian, &temp16); err != nil {
		return err
	}
	o.SrcRect.Max.Y = int(temp16)
	if err := binary.Read(r, binary.BigEndian, &temp16); err != nil {
		return err
	}
	o.SrcRect.Max.X = int(temp16)

	// reserved
	var temp32 uint32
	if err := binary.Read(r, binary.BigEndian, &temp32); err != nil {
		return err
	}

	return nil
}
func (o *OpHeader) Dump() {
	fmt.Printf("  Op Header: %d.%d,%d.%d %+v\n",
		o.ResH>>16,
		o.ResH&0xffff,
		o.ResV>>16,
		o.ResV&0xffff,
		o.SrcRect)
}

type OpClip struct {
	Opcode uint16
	Rect   image.Rectangle
}

func (o *OpClip) Parse(r *io.SectionReader) error {
	var temp16 uint16

	// check size
	if err := binary.Read(r, binary.BigEndian, &temp16); err != nil {
		return err
	}
	if temp16 != 10 {
		return fmt.Errorf("invalid size: %04x", temp16)
	}

	// Rect
	if err := binary.Read(r, binary.BigEndian, &temp16); err != nil {
		return err
	}
	o.Rect.Min.Y = int(temp16)
	if err := binary.Read(r, binary.BigEndian, &temp16); err != nil {
		return err
	}
	o.Rect.Min.X = int(temp16)
	if err := binary.Read(r, binary.BigEndian, &temp16); err != nil {
		return err
	}
	o.Rect.Max.Y = int(temp16)
	if err := binary.Read(r, binary.BigEndian, &temp16); err != nil {
		return err
	}
	o.Rect.Max.X = int(temp16)

	return nil
}
func (o *OpClip) Dump() {
	fmt.Printf("  Op Clip: %+v\n", o.Rect)
}

type OpQTcomp struct {
	Opcode uint16
	Size   uint32

	reader *io.SectionReader
}

func (o *OpQTcomp) Parse(r *io.SectionReader) error {
	// Size
	if err := binary.Read(r, binary.BigEndian, &o.Size); err != nil {
		return err
	}

	// reader for QuickTime compressed pictures
	off, err := r.Seek(int64(o.Size), io.SeekCurrent)
	if err != nil {
		return err
	}
	o.reader = io.NewSectionReader(r, off-int64(o.Size), int64(o.Size))

	return nil
}
func (o *OpQTcomp) Dump() {
	fmt.Printf("  Op QTcomp: %+v\n", o)
}
func (o *OpQTcomp) DumpTo(w io.Writer) (int64, error) {
	// atom
	var dataSize uint32
	_, err := o.reader.Seek(68+32+2+2+4+4, io.SeekStart)
	if err != nil {
		return 0, err
	}
	if err := binary.Read(o.reader, binary.BigEndian, &dataSize); err != nil {
		return 0, err
	}

	// jfif
	_, err = o.reader.Seek(68+0x56, io.SeekStart)
	if err != nil {
		return 0, err
	}
	return io.CopyN(w, o.reader, int64(dataSize))
}

type OpPackBitsRect struct {
	Opcode uint16

	RowBytes uint16
	Bounds   image.Rectangle
	SrcRect  image.Rectangle
	DstRect  image.Rectangle
	Mode     uint16

	pixData  [][]byte
	unpacked [][]byte
}

func (o *OpPackBitsRect) Parse(r *io.SectionReader) error {
	var temp16 uint16

	// Row bytes
	if err := binary.Read(r, binary.BigEndian, &o.RowBytes); err != nil {
		return err
	}
	if o.RowBytes < 8 {
		return errors.New("data is unpacked")
	}
	if o.RowBytes&0x8000 != 0 {
		return errors.New("containing multiple bits per pixel")
	}

	// Bounds
	if err := binary.Read(r, binary.BigEndian, &temp16); err != nil {
		return err
	}
	o.Bounds.Min.Y = int(temp16)
	if err := binary.Read(r, binary.BigEndian, &temp16); err != nil {
		return err
	}
	o.Bounds.Min.X = int(temp16)
	if err := binary.Read(r, binary.BigEndian, &temp16); err != nil {
		return err
	}
	o.Bounds.Max.Y = int(temp16)
	if err := binary.Read(r, binary.BigEndian, &temp16); err != nil {
		return err
	}
	o.Bounds.Max.X = int(temp16)

	// Src rect
	if err := binary.Read(r, binary.BigEndian, &temp16); err != nil {
		return err
	}
	o.SrcRect.Min.Y = int(temp16)
	if err := binary.Read(r, binary.BigEndian, &temp16); err != nil {
		return err
	}
	o.SrcRect.Min.X = int(temp16)
	if err := binary.Read(r, binary.BigEndian, &temp16); err != nil {
		return err
	}
	o.SrcRect.Max.Y = int(temp16)
	if err := binary.Read(r, binary.BigEndian, &temp16); err != nil {
		return err
	}
	o.SrcRect.Max.X = int(temp16)

	// Dst Rect
	if err := binary.Read(r, binary.BigEndian, &temp16); err != nil {
		return err
	}
	o.DstRect.Min.Y = int(temp16)
	if err := binary.Read(r, binary.BigEndian, &temp16); err != nil {
		return err
	}
	o.DstRect.Min.X = int(temp16)
	if err := binary.Read(r, binary.BigEndian, &temp16); err != nil {
		return err
	}
	o.DstRect.Max.Y = int(temp16)
	if err := binary.Read(r, binary.BigEndian, &temp16); err != nil {
		return err
	}
	o.DstRect.Max.X = int(temp16)

	// Mode
	if err := binary.Read(r, binary.BigEndian, &o.Mode); err != nil {
		return err
	}

	// PixData
	o.pixData = make([][]byte, o.Bounds.Dy())
	for y := 0; y < o.Bounds.Dy(); y++ {
		var count int
		if o.RowBytes > 250 {
			if err := binary.Read(r, binary.BigEndian, &temp16); err != nil {
				return err
			}
			count = int(temp16)
		} else {
			var temp8 byte
			if err := binary.Read(r, binary.BigEndian, &temp8); err != nil {
				return err
			}
			count = int(temp8)
		}

		o.pixData[y] = make([]byte, count)
		_, err := io.ReadFull(r, o.pixData[y])
		if err != nil {
			return nil
		}
	}

	// unpack
	o.unpacked = make([][]byte, o.Bounds.Dy())
	for y := 0; y < o.Bounds.Dy(); y++ {
		packed := o.pixData[y]
		unpacked := make([]byte, o.RowBytes)
		o.unpacked[y] = unpacked
		for len(packed) > 0 {
			l := packed[0]
			packed = packed[1:]
			if l&0x80 == 0 {
				l += 1
				copy(unpacked[0:l], packed[0:l])
				packed = packed[l:]
				unpacked = unpacked[l:]
			} else {
				l = 255 - l + 1 + 1
				d := packed[0]
				for i := byte(0); i < l; i++ {
					unpacked[i] = d
				}
				packed = packed[1:]
				unpacked = unpacked[l:]
			}
		}
	}

	return nil
}
func (o *OpPackBitsRect) Dump() {
	fmt.Printf("  Op PackBitsRect: RowBytes=%d, Bounds=%+v, SrcRect=%+v, DstRect=%+v, Mode=%d\n",
		o.RowBytes,
		o.Bounds,
		o.SrcRect,
		o.DstRect,
		o.Mode,
	)
	for i, row := range o.unpacked {
		fmt.Printf("    %02d: ", i)
		for _, d := range row {
			for b := 7; b >= 0; b-- {
				p := d & (1 << b)
				if p != 0 {
					fmt.Printf("@")
				} else {
					fmt.Printf(" ")
				}
			}
		}
		fmt.Println("")
	}
}

type OpEndPic struct {
	Opcode uint16
}

func (o *OpEndPic) Parse(r *io.SectionReader) error {
	// no data
	return nil
}
func (o *OpEndPic) Dump() {
	fmt.Printf("  Op EndPic\n")
}

func NewOp(opcode uint16, r *io.SectionReader) (Operator, error) {
	var op Operator
	switch opcode {
	case Version:
		op = &OpVersion{Opcode: opcode}
	case Header:
		op = &OpHeader{Opcode: opcode}
	case Clip:
		op = &OpClip{Opcode: opcode}
	case QTcomp:
		op = &OpQTcomp{Opcode: opcode}
	case PackBitsRect:
		op = &OpPackBitsRect{Opcode: opcode}
	case EndPic:
		op = &OpEndPic{Opcode: opcode}
	default:
		off, _ := r.Seek(0, io.SeekCurrent)
		return nil, fmt.Errorf("not implemented: opcode=%04x, off=%08x", opcode, off)
	}

	err := op.Parse(r)
	return op, err
}
