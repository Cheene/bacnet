package encoding

import (
	"bytes"
	"encoding/binary"
	"math"
	"testing"
)

func TestReadUint16(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected uint16
	}{
		{"valid", []byte{0x12, 0x34}, 0x1234},
		{"zero", []byte{0x00, 0x00}, 0},
		{"max", []byte{0xFF, 0xFF}, 0xFFFF},
		{"short", []byte{0x12}, 0},
		{"empty", []byte{}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := readUint16(tt.input)
			if result != tt.expected {
				t.Errorf("readUint16(%v) = %d, expected %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestReadUint32(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected uint32
	}{
		{"valid", []byte{0x12, 0x34, 0x56, 0x78}, 0x12345678},
		{"zero", []byte{0x00, 0x00, 0x00, 0x00}, 0},
		{"max", []byte{0xFF, 0xFF, 0xFF, 0xFF}, 0xFFFFFFFF},
		{"short", []byte{0x12}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := readUint32(tt.input)
			if result != tt.expected {
				t.Errorf("readUint32(%v) = %d, expected %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestReadInt32(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected int32
	}{
		{"positive", []byte{0x00, 0x00, 0x00, 0x7F}, 127},
		{"negative", []byte{0xFF, 0xFF, 0xFF, 0xFF}, -1},
		{"zero", []byte{0x00, 0x00, 0x00, 0x00}, 0},
		{"min", []byte{0x80, 0x00, 0x00, 0x00}, math.MinInt32},
		{"max", []byte{0x7F, 0xFF, 0xFF, 0xFF}, math.MaxInt32},
		{"short", []byte{0x12}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := readInt32(tt.input)
			if result != tt.expected {
				t.Errorf("readInt32(%v) = %d, expected %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestReadFloat32(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected float32
	}{
		{"positive", []byte{0x41, 0x48, 0x00, 0x00}, 12.5},
		{"zero", []byte{0x00, 0x00, 0x00, 0x00}, 0},
		{"negative", []byte{0xC1, 0x48, 0x00, 0x00}, -12.5},
		{"short", []byte{0x12}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := readFloat32(tt.input)
			if result != tt.expected {
				t.Errorf("readFloat32(%v) = %f, expected %f", tt.input, result, tt.expected)
			}
		})
	}
}

func TestReadFloat64(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected float64
	}{
		{"positive", []byte{0x40, 0x25, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, 10.5},
		{"zero", []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, 0},
		{"negative", []byte{0xC0, 0x25, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, -10.5},
		{"short", []byte{0x12}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := readFloat64(tt.input)
			if result != tt.expected {
				t.Errorf("readFloat64(%v) = %f, expected %f", tt.input, result, tt.expected)
			}
		})
	}
}

func TestWriteUint16(t *testing.T) {
	tests := []struct {
		name     string
		value    uint16
		expected []byte
	}{
		{"value", 0x1234, []byte{0x12, 0x34}},
		{"zero", 0, []byte{0x00, 0x00}},
		{"max", 0xFFFF, []byte{0xFF, 0xFF}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := make([]byte, 2)
			writeUint16(buf, tt.value)
			if buf[0] != tt.expected[0] || buf[1] != tt.expected[1] {
				t.Errorf("writeUint16(%d) = %v, expected %v", tt.value, buf, tt.expected)
			}
		})
	}
}

func TestWriteUint32(t *testing.T) {
	tests := []struct {
		name     string
		value    uint32
		expected []byte
	}{
		{"value", 0x12345678, []byte{0x12, 0x34, 0x56, 0x78}},
		{"zero", 0, []byte{0x00, 0x00, 0x00, 0x00}},
		{"max", 0xFFFFFFFF, []byte{0xFF, 0xFF, 0xFF, 0xFF}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := make([]byte, 4)
			writeUint32(buf, tt.value)
			for i, b := range tt.expected {
				if buf[i] != b {
					t.Errorf("writeUint32(%d) = %v, expected %v", tt.value, buf, tt.expected)
					break
				}
			}
		})
	}
}

func TestWriteInt32(t *testing.T) {
	tests := []struct {
		name     string
		value    int32
		expected []byte
	}{
		{"positive", 127, []byte{0x00, 0x00, 0x00, 0x7F}},
		{"negative", -1, []byte{0xFF, 0xFF, 0xFF, 0xFF}},
		{"zero", 0, []byte{0x00, 0x00, 0x00, 0x00}},
		{"min", math.MinInt32, []byte{0x80, 0x00, 0x00, 0x00}},
		{"max", math.MaxInt32, []byte{0x7F, 0xFF, 0xFF, 0xFF}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := make([]byte, 4)
			writeInt32(buf, tt.value)
			for i, b := range tt.expected {
				if buf[i] != b {
					t.Errorf("writeInt32(%d) = %v, expected %v", tt.value, buf, tt.expected)
					break
				}
			}
		})
	}
}

func TestWriteFloat32(t *testing.T) {
	tests := []struct {
		name     string
		value    float32
		expected []byte
	}{
		{"positive", 12.5, []byte{0x41, 0x48, 0x00, 0x00}},
		{"zero", 0, []byte{0x00, 0x00, 0x00, 0x00}},
		{"negative", -12.5, []byte{0xC1, 0x48, 0x00, 0x00}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := make([]byte, 4)
			writeFloat32(buf, tt.value)
			for i, b := range tt.expected {
				if buf[i] != b {
					t.Errorf("writeFloat32(%f) = %v, expected %v", tt.value, buf, tt.expected)
					break
				}
			}
		})
	}
}

func TestWriteFloat64(t *testing.T) {
	tests := []struct {
		name     string
		value    float64
		expected []byte
	}{
		{"positive", 10.5, []byte{0x40, 0x25, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
		{"zero", 0, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
		{"negative", -10.5, []byte{0xC0, 0x25, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := make([]byte, 8)
			writeFloat64(buf, tt.value)
			for i, b := range tt.expected {
				if buf[i] != b {
					t.Errorf("writeFloat64(%f) = %v, expected %v", tt.value, buf, tt.expected)
					break
				}
			}
		})
	}
}

func TestWriteShortBuffer(t *testing.T) {
	buf := make([]byte, 1)
	writeUint16(buf, 0x1234)
	buf = make([]byte, 3)
	writeUint32(buf, 0x12345678)
	writeInt32(buf, 123456)
	buf = make([]byte, 7)
	writeFloat64(buf, 10.5)
}

func TestAlignedRead(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected uint16
	}{
		{"valid", []byte{0x12, 0x34}, 0x1234},
		{"zero", []byte{0x00, 0x00}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bytes.NewReader(tt.input)
			var result uint16
			err := alignedRead(r, binary.BigEndian, &result)
			if err != nil {
				t.Errorf("alignedRead error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("alignedRead(%v) = %d, expected %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestAlignedWrite(t *testing.T) {
	tests := []struct {
		name     string
		value    uint16
		expected []byte
	}{
		{"value", 0x1234, []byte{0x12, 0x34}},
		{"zero", 0, []byte{0x00, 0x00}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			err := alignedWrite(buf, binary.BigEndian, tt.value)
			if err != nil {
				t.Errorf("alignedWrite error: %v", err)
			}
			result := buf.Bytes()
			if !bytes.Equal(result, tt.expected) {
				t.Errorf("alignedWrite(%d) = %v, expected %v", tt.value, result, tt.expected)
			}
		})
	}
}
