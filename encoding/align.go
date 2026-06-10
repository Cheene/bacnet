package encoding

import (
	"encoding/binary"
	"io"
	"math"
)

// alignedRead performs binary.Read with alignment-safe handling
// This ensures proper alignment for ARMv7 architecture requirements
func alignedRead(r io.Reader, order binary.ByteOrder, data interface{}) error {
	return binary.Read(r, order, data)
}

// alignedWrite performs binary.Write with alignment-safe handling
// This ensures proper alignment for ARMv7 architecture requirements
func alignedWrite(w io.Writer, order binary.ByteOrder, data interface{}) error {
	return binary.Write(w, order, data)
}

// readUint16 reads a uint16 from buffer in an alignment-safe manner
// ARMv7 requires 16-bit data to be aligned to 2-byte boundary
func readUint16(b []byte) uint16 {
	if len(b) < 2 {
		return 0
	}
	return uint16(b[0])<<8 | uint16(b[1])
}

// readUint32 reads a uint32 from buffer in an alignment-safe manner
// ARMv7 requires 32-bit data to be aligned to 4-byte boundary
func readUint32(b []byte) uint32 {
	if len(b) < 4 {
		return 0
	}
	return uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
}

// readInt32 reads an int32 from buffer in an alignment-safe manner
func readInt32(b []byte) int32 {
	if len(b) < 4 {
		return 0
	}
	return int32(uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3]))
}

// readFloat32 reads a float32 from buffer in an alignment-safe manner
// ARMv7 requires float32 to be aligned to 4-byte boundary
func readFloat32(b []byte) float32 {
	if len(b) < 4 {
		return 0
	}
	return math.Float32frombits(readUint32(b))
}

// readFloat64 reads a float64 from buffer in an alignment-safe manner
// ARMv7 requires float64 to be aligned to 8-byte boundary
func readFloat64(b []byte) float64 {
	if len(b) < 8 {
		return 0
	}
	// Read as two uint32 values and combine
	high := readUint32(b[:4])
	low := readUint32(b[4:8])
	bits := (uint64(high) << 32) | uint64(low)
	return math.Float64frombits(bits)
}

// writeUint16 writes a uint16 to buffer in an alignment-safe manner
func writeUint16(b []byte, v uint16) {
	if len(b) >= 2 {
		b[0] = byte(v >> 8)
		b[1] = byte(v)
	}
}

// writeUint32 writes a uint32 to buffer in an alignment-safe manner
func writeUint32(b []byte, v uint32) {
	if len(b) >= 4 {
		b[0] = byte(v >> 24)
		b[1] = byte(v >> 16)
		b[2] = byte(v >> 8)
		b[3] = byte(v)
	}
}

// writeInt32 writes an int32 to buffer in an alignment-safe manner
func writeInt32(b []byte, v int32) {
	writeUint32(b, uint32(v))
}

// writeFloat32 writes a float32 to buffer in an alignment-safe manner
func writeFloat32(b []byte, v float32) {
	writeUint32(b, math.Float32bits(v))
}

// writeFloat64 writes a float64 to buffer in an alignment-safe manner
func writeFloat64(b []byte, v float64) {
	bits := math.Float64bits(v)
	if len(b) >= 8 {
		b[0] = byte(bits >> 56)
		b[1] = byte(bits >> 48)
		b[2] = byte(bits >> 40)
		b[3] = byte(bits >> 32)
		b[4] = byte(bits >> 24)
		b[5] = byte(bits >> 16)
		b[6] = byte(bits >> 8)
		b[7] = byte(bits)
	}
}
