package tiff

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

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
	Tag     uint16
	IFDType uint16
	Count   uint32
	Offset  uint32

	Values []interface{}

	// for debug
	globalOffset int64

	// cache
	elmSize int64
}

// String makes IFDEntry satisfy the Stringer interface.
func (e *IFDEntry) String() string {
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("    Tag: %xh\n", e.Tag))
	buf.WriteString(fmt.Sprintf("    Type: %d\n", e.IFDType))
	buf.WriteString(fmt.Sprintf("    Count: %d\n", e.Count))
	buf.WriteString(fmt.Sprintf("    Offset: 0x%08x (global: 0x%08x)\n", e.Offset, e.globalOffset))
	buf.WriteString(fmt.Sprintf("    Value: %+v\n", e.Values))

	return buf.String()
}

func parseIFD(rs io.ReadSeeker, byteOrder binary.ByteOrder, globalOffset int64) (int64, []*IFDEntry, error) {
	var num uint16
	if err := binary.Read(rs, byteOrder, &num); err != nil {
		return 0, nil, err
	}

	entries := make([]*IFDEntry, 0, num)
	for i := num; i > 0; i-- {
		entry := &IFDEntry{}

		if err := binary.Read(rs, byteOrder, &entry.Tag); err != nil {
			return 0, entries, err
		}
		if err := binary.Read(rs, byteOrder, &entry.IFDType); err != nil {
			return 0, entries, err
		}
		if err := binary.Read(rs, byteOrder, &entry.Count); err != nil {
			return 0, entries, err
		}
		// Offset or Value
		totalBytes := entry.elementSize() * int64(entry.Count)
		if totalBytes > 4 {
			// Offset
			if err := binary.Read(rs, byteOrder, &entry.Offset); err != nil {
				return 0, entries, err
			}
			entry.Values = nil
		} else {
			// Value
			entry.Offset = 0
			if err := entry.parseValue4bytes(rs, byteOrder); err != nil {
				return 0, entries, err
			}
		}
		entry.globalOffset = globalOffset + int64(entry.Offset)

		entries = append(entries, entry)
	}

	var offsetNext uint32
	if err := binary.Read(rs, byteOrder, &offsetNext); err != nil {
		return 0, entries, err
	}

	return int64(offsetNext), entries, nil
}

func (e *IFDEntry) elementSize() int64 {
	if e.elmSize != 0 {
		return e.elmSize
	}

	switch e.IFDType {
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

	e.Values = make([]interface{}, 0, e.Count)

	switch e.IFDType {
	case BYTE:
		var value uint8
		for i := e.Count; i > 0; i-- {
			if err := binary.Read(rs, byteOrder, &value); err != nil {
				return err
			}
			e.Values = append(e.Values, value)
		}
	case ASCII:
		var value byte
		for i := e.Count; i > 0; i-- {
			if err := binary.Read(rs, byteOrder, &value); err != nil {
				return err
			}
			e.Values = append(e.Values, value)
		}
	case SHORT:
		var value uint16
		for i := e.Count; i > 0; i-- {
			if err := binary.Read(rs, byteOrder, &value); err != nil {
				return err
			}
			e.Values = append(e.Values, value)
		}
	case LONG:
		var value uint32
		for i := e.Count; i > 0; i-- {
			if err := binary.Read(rs, byteOrder, &value); err != nil {
				return err
			}
			e.Values = append(e.Values, value)
		}
	case RATIONAL:
		// over 4bytes
		e.Values = nil
	case SBYTE:
		var value int8
		for i := e.Count; i > 0; i-- {
			if err := binary.Read(rs, byteOrder, &value); err != nil {
				return err
			}
			e.Values = append(e.Values, value)
		}
	case UNDEFINED:
		var value byte
		for i := e.Count; i > 0; i-- {
			if err := binary.Read(rs, byteOrder, &value); err != nil {
				return err
			}
			e.Values = append(e.Values, value)
		}
	case SSHORT:
		var value int16
		for i := e.Count; i > 0; i-- {
			if err := binary.Read(rs, byteOrder, &value); err != nil {
				return err
			}
			e.Values = append(e.Values, value)
		}
	case SLONG:
		var value int32
		for i := e.Count; i > 0; i-- {
			if err := binary.Read(rs, byteOrder, &value); err != nil {
				return err
			}
			e.Values = append(e.Values, value)
		}
	case SRATIONAL:
		// over 4bytes
		e.Values = nil
	default:
		e.Values = nil
	}

	_, err = rs.Seek(pos+4, io.SeekStart)
	return err

	//return nil
}
