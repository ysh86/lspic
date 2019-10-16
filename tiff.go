package lsjpeg

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

func parseTiffHeader(rs io.ReadSeeker) (int64, int64, binary.ByteOrder, error) {
	offsetHeader, err := rs.Seek(0, io.SeekCurrent)
	if err != nil {
		return 0, 0, nil, err
	}

	var endian uint16
	if err := binary.Read(rs, binary.BigEndian, &endian); err != nil {
		return 0, offsetHeader, nil, err
	}
	var byteOrder binary.ByteOrder
	if endian == 0x4949 {
		byteOrder = binary.LittleEndian
	} else if endian == 0x4d4d {
		byteOrder = binary.BigEndian
	} else {
		return 0, offsetHeader, nil, errors.New("invalid byte order")
	}

	var value42 uint16
	if err := binary.Read(rs, byteOrder, &value42); err != nil {
		return 0, offsetHeader, byteOrder, err
	}
	if value42 != 0x002a {
		return 0, offsetHeader, byteOrder, errors.New("invalid 42")
	}

	var offsetNext uint32
	if err := binary.Read(rs, byteOrder, &offsetNext); err != nil {
		return 0, offsetHeader, byteOrder, err
	}

	return int64(offsetNext), offsetHeader, byteOrder, nil
}

func parseTiffIfd(rs io.ReadSeeker, byteOrder binary.ByteOrder) (int64, []*IFDEntry, error) {
	var num uint16
	if err := binary.Read(rs, byteOrder, &num); err != nil {
		return 0, nil, err
	}

	entries := make([]*IFDEntry, 0, num)
	for i := num; i > 0; i-- {
		entry := &IFDEntry{}

		if err := binary.Read(rs, byteOrder, &entry.tag); err != nil {
			return 0, entries, err
		}
		if err := binary.Read(rs, byteOrder, &entry.ifdType); err != nil {
			return 0, entries, err
		}
		if err := binary.Read(rs, byteOrder, &entry.count); err != nil {
			return 0, entries, err
		}
		// Offset or Value
		totalBytes := entry.elementSize() * int64(entry.count)
		if totalBytes > 4 {
			// Offset
			if err := binary.Read(rs, byteOrder, &entry.offset); err != nil {
				return 0, entries, err
			}
			entry.values = nil
		} else {
			// Value
			entry.offset = 0
			if err := entry.parseValue4bytes(rs, byteOrder); err != nil {
				return 0, entries, err
			}
		}

		entries = append(entries, entry)
	}

	var offsetNext uint32
	if err := binary.Read(rs, byteOrder, &offsetNext); err != nil {
		return 0, entries, err
	}

	return int64(offsetNext), entries, nil
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

	return buf.String()
}
