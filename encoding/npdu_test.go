package encoding

import (
	"testing"

	"github.com/anviod/bacnet/btypes"
	"github.com/anviod/bacnet/btypes/ndpu"
)

func TestEncoderNPDU(t *testing.T) {
	tests := []struct {
		name string
		npdu btypes.NPDU
	}{
		{"basic", btypes.NPDU{Version: 1}},
		{"with destination", btypes.NPDU{
			Version:     1,
			Destination: &btypes.Address{Net: 123, Adr: []byte{0x01, 0x02, 0x03}},
		}},
		{"with source", btypes.NPDU{
			Version: 1,
			Source:  &btypes.Address{Net: 456, Adr: []byte{0x04, 0x05}},
		}},
		{"destination with Mac", btypes.NPDU{
			Version:     1,
			Destination: &btypes.Address{Net: 789, Mac: []byte{0x01, 0x02, 0x03, 0x04, 0x05}, Id: 0xAA},
		}},
		{"network layer message", btypes.NPDU{
			Version:                 1,
			IsNetworkLayerMessage:   true,
			NetworkLayerMessageType: 0x12,
		}},
		{"network layer message with vendor id", btypes.NPDU{
			Version:                 1,
			IsNetworkLayerMessage:   true,
			NetworkLayerMessageType: 0x81,
			VendorId:                0xABCD,
		}},
		{"expecting reply", btypes.NPDU{
			Version:        1,
			ExpectingReply: true,
			Priority:       btypes.Urgent,
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewEncoder()
			e.NPDU(&tt.npdu)
			if err := e.Error(); err != nil {
				t.Fatal(err)
			}
			b := e.Bytes()
			d := NewDecoder(b)

			var out btypes.NPDU
			_, err := d.NPDU(&out)
			if err != nil {
				t.Fatal(err)
			}

			if out.Version != tt.npdu.Version {
				t.Errorf("Version = %d, expected %d", out.Version, tt.npdu.Version)
			}
			if out.IsNetworkLayerMessage != tt.npdu.IsNetworkLayerMessage {
				t.Errorf("IsNetworkLayerMessage = %v, expected %v", out.IsNetworkLayerMessage, tt.npdu.IsNetworkLayerMessage)
			}
			if out.NetworkLayerMessageType != tt.npdu.NetworkLayerMessageType {
				t.Errorf("NetworkLayerMessageType = %d, expected %d", out.NetworkLayerMessageType, tt.npdu.NetworkLayerMessageType)
			}
		})
	}
}

func TestDecoderAddress(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected btypes.Address
	}{
		{"valid address", []byte{0x00, 0x7B, 0x03, 0x01, 0x02, 0x03},
			btypes.Address{Net: 123, Len: 3, Adr: []byte{0x01, 0x02, 0x03}}},
		{"empty address", []byte{0x00, 0x00, 0x00},
			btypes.Address{Net: 0, Len: 0, Adr: []byte{}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDecoder(tt.input)
			a := &btypes.Address{}
			d.Address(a)
			if a.Net != tt.expected.Net {
				t.Errorf("Net = %d, expected %d", a.Net, tt.expected.Net)
			}
			if a.Len != tt.expected.Len {
				t.Errorf("Len = %d, expected %d", a.Len, tt.expected.Len)
			}
		})
	}
}

func TestNPDUMetadata(t *testing.T) {
	t.Run("SetPriority", func(t *testing.T) {
		var meta NPDUMetadata
		meta.SetPriority(btypes.Urgent)
		if meta.Priority() != btypes.Urgent {
			t.Errorf("Priority() = %d, expected %d", meta.Priority(), btypes.Urgent)
		}
	})

	t.Run("HasDestination", func(t *testing.T) {
		var meta NPDUMetadata
		if meta.HasDestination() {
			t.Error("HasDestination() = true, expected false")
		}
		meta.SetDestination(true)
		if !meta.HasDestination() {
			t.Error("HasDestination() = false, expected true")
		}
	})

	t.Run("HasSource", func(t *testing.T) {
		var meta NPDUMetadata
		if meta.HasSource() {
			t.Error("HasSource() = true, expected false")
		}
		meta.SetSource(true)
		if !meta.HasSource() {
			t.Error("HasSource() = false, expected true")
		}
	})

	t.Run("ExpectingReply", func(t *testing.T) {
		var meta NPDUMetadata
		if meta.ExpectingReply() {
			t.Error("ExpectingReply() = true, expected false")
		}
		meta.SetExpectingReply(true)
		if !meta.ExpectingReply() {
			t.Error("ExpectingReply() = false, expected true")
		}
	})

	t.Run("IsNetworkLayerMessage", func(t *testing.T) {
		var meta NPDUMetadata
		if meta.IsNetworkLayerMessage() {
			t.Error("IsNetworkLayerMessage() = true, expected false")
		}
		meta.SetNetworkLayerMessage(true)
		if !meta.IsNetworkLayerMessage() {
			t.Error("IsNetworkLayerMessage() = false, expected true")
		}
	})
}

func TestDecoderNPDUNetworkIs(t *testing.T) {
	input := []byte{0x01, 0x80, byte(ndpu.NetworkIs), 0x00, 0x7B}
	d := NewDecoder(input)
	n := &btypes.NPDU{}
	d.NPDU(n)
	if n.NetworkLayerMessageType != ndpu.NetworkIs {
		t.Errorf("NetworkLayerMessageType = %d, expected %d", n.NetworkLayerMessageType, ndpu.NetworkIs)
	}
	if n.Source == nil || n.Source.Net != 123 {
		t.Errorf("Source.Net = %d, expected 123", n.Source.Net)
	}
}

func TestDecoderNPDUIamRouterToNetwork(t *testing.T) {
	input := []byte{0x01, 0x80, byte(ndpu.IamRouterToNetwork), 0x00, 0x7B}
	d := NewDecoder(input)
	n := &btypes.NPDU{}
	addr, _ := d.NPDU(n)
	if n.NetworkLayerMessageType != ndpu.IamRouterToNetwork {
		t.Errorf("NetworkLayerMessageType = %d, expected %d", n.NetworkLayerMessageType, ndpu.IamRouterToNetwork)
	}
	if len(addr) > 0 {
		t.Logf("Router addresses: %v", addr)
	}
}

func TestNPDUWithHopCount(t *testing.T) {
	input := []byte{0x01, 0x20, 0x00, 0x7B, 0x02, 0x01, 0x02, 0x05}
	d := NewDecoder(input)
	n := &btypes.NPDU{}
	d.NPDU(n)
	if n.HopCount != 5 {
		t.Errorf("HopCount = %d, expected 5", n.HopCount)
	}
}

func TestNPDUWithoutDestinationHopCount(t *testing.T) {
	input := []byte{0x01, 0x00}
	d := NewDecoder(input)
	n := &btypes.NPDU{}
	d.NPDU(n)
	if n.HopCount != 0 {
		t.Errorf("HopCount = %d, expected 0", n.HopCount)
	}
}

func TestNPDUWithSourceAndDestination(t *testing.T) {
	n := btypes.NPDU{
		Version:     1,
		Destination: &btypes.Address{Net: 100, Adr: []byte{0x01, 0x02}},
		Source:      &btypes.Address{Net: 200, Adr: []byte{0x03, 0x04}},
	}

	e := NewEncoder()
	e.NPDU(&n)
	if err := e.Error(); err != nil {
		t.Fatal(err)
	}
	b := e.Bytes()
	d := NewDecoder(b)

	var out btypes.NPDU
	_, err := d.NPDU(&out)
	if err != nil {
		t.Fatal(err)
	}

	if out.Destination == nil || out.Destination.Net != 100 {
		t.Errorf("Destination.Net = %d, expected 100", out.Destination.Net)
	}
	if out.Source == nil || out.Source.Net != 200 {
		t.Errorf("Source.Net = %d, expected 200", out.Source.Net)
	}
}

func TestDecoderNPDUNetworkLayerMessageWithVendorId(t *testing.T) {
	input := []byte{0x01, 0x80, 0x81, 0xAB, 0xCD}
	d := NewDecoder(input)
	n := &btypes.NPDU{}
	d.NPDU(n)
	if n.NetworkLayerMessageType != 0x81 {
		t.Errorf("NetworkLayerMessageType = %d, expected 0x81", n.NetworkLayerMessageType)
	}
	if n.VendorId != 0xABCD {
		t.Errorf("VendorId = %d, expected 0xABCD", n.VendorId)
	}
}

func TestDecoderNPDUWithSourceOnly(t *testing.T) {
	input := []byte{0x01, 0x08, 0x00, 0xC8, 0x02, 0x01, 0x02}
	d := NewDecoder(input)
	n := &btypes.NPDU{}
	d.NPDU(n)
	if n.Source == nil || n.Source.Net != 200 {
		t.Errorf("Source.Net = %d, expected 200", n.Source.Net)
	}
	if n.HopCount != 0 {
		t.Errorf("HopCount = %d, expected 0", n.HopCount)
	}
}

func TestNPDUBoundaryConditions(t *testing.T) {
	t.Run("max network address", func(t *testing.T) {
		n := btypes.NPDU{
			Version:     1,
			Destination: &btypes.Address{Net: 0xFFFF, Adr: []byte{0x01}},
		}
		e := NewEncoder()
		e.NPDU(&n)
		b := e.Bytes()
		d := NewDecoder(b)
		var out btypes.NPDU
		d.NPDU(&out)
		if out.Destination.Net != 0xFFFF {
			t.Errorf("Destination.Net = %d, expected 0xFFFF", out.Destination.Net)
		}
	})

	t.Run("max address length", func(t *testing.T) {
		addrBytes := make([]byte, 255)
		for i := range addrBytes {
			addrBytes[i] = byte(i)
		}
		n := btypes.NPDU{
			Version:     1,
			Destination: &btypes.Address{Net: 1, Adr: addrBytes},
		}
		e := NewEncoder()
		e.NPDU(&n)
		b := e.Bytes()
		d := NewDecoder(b)
		var out btypes.NPDU
		d.NPDU(&out)
		if len(out.Destination.Adr) != 255 {
			t.Errorf("Address length = %d, expected 255", len(out.Destination.Adr))
		}
	})

	t.Run("empty address bytes", func(t *testing.T) {
		n := btypes.NPDU{
			Version:     1,
			Destination: &btypes.Address{Net: 1, Adr: []byte{}},
		}
		e := NewEncoder()
		e.NPDU(&n)
		b := e.Bytes()
		d := NewDecoder(b)
		var out btypes.NPDU
		d.NPDU(&out)
		if out.Destination.Len != 0 {
			t.Errorf("Address Len = %d, expected 0", out.Destination.Len)
		}
	})

	t.Run("nil destination", func(t *testing.T) {
		n := btypes.NPDU{
			Version:     1,
			Destination: nil,
			Source:      &btypes.Address{Net: 100, Adr: []byte{0x01}},
		}
		e := NewEncoder()
		e.NPDU(&n)
		b := e.Bytes()
		d := NewDecoder(b)
		var out btypes.NPDU
		d.NPDU(&out)
		if out.Destination != nil {
			t.Error("Destination should be nil")
		}
		if out.Source == nil || out.Source.Net != 100 {
			t.Errorf("Source.Net = %d, expected 100", out.Source.Net)
		}
	})

	t.Run("nil source", func(t *testing.T) {
		n := btypes.NPDU{
			Version:     1,
			Destination: &btypes.Address{Net: 100, Adr: []byte{0x01}},
			Source:      nil,
		}
		e := NewEncoder()
		e.NPDU(&n)
		b := e.Bytes()
		d := NewDecoder(b)
		var out btypes.NPDU
		d.NPDU(&out)
		if out.Source != nil {
			t.Error("Source should be nil")
		}
		if out.Destination == nil || out.Destination.Net != 100 {
			t.Errorf("Destination.Net = %d, expected 100", out.Destination.Net)
		}
	})

	t.Run("all priority levels", func(t *testing.T) {
		priorities := []btypes.NPDUPriority{btypes.Normal, btypes.Urgent, btypes.CriticalEquipment, btypes.LifeSafety}
		for _, p := range priorities {
			n := btypes.NPDU{
				Version:  1,
				Priority: p,
			}
			e := NewEncoder()
			e.NPDU(&n)
			b := e.Bytes()
			d := NewDecoder(b)
			var out btypes.NPDU
			d.NPDU(&out)
			if out.Priority != p {
				t.Errorf("Priority = %d, expected %d", out.Priority, p)
			}
		}
	})

	t.Run("destination with empty Adr but valid Mac", func(t *testing.T) {
		n := btypes.NPDU{
			Version:     1,
			Destination: &btypes.Address{Net: 100, Mac: []byte{0x01, 0x02, 0x03, 0x04}, Id: 0x55},
		}
		e := NewEncoder()
		e.NPDU(&n)
		b := e.Bytes()
		if len(b) < 8 {
			t.Errorf("Encoded length = %d, expected at least 8", len(b))
		}
	})

	t.Run("destination with short Mac fallback", func(t *testing.T) {
		n := btypes.NPDU{
			Version:     1,
			Destination: &btypes.Address{Net: 100, Mac: []byte{0x01, 0x02}, Id: 0x55},
		}
		e := NewEncoder()
		e.NPDU(&n)
		b := e.Bytes()
		d := NewDecoder(b)
		var out btypes.NPDU
		d.NPDU(&out)
		if out.Destination == nil || out.Destination.Net != 100 {
			t.Errorf("Destination.Net = %d, expected 100", out.Destination.Net)
		}
	})

	t.Run("all metadata flags set", func(t *testing.T) {
		n := btypes.NPDU{
			Version:               1,
			IsNetworkLayerMessage: true,
			ExpectingReply:        true,
			Priority:              btypes.Urgent,
			Destination:           &btypes.Address{Net: 100, Adr: []byte{0x01}},
			Source:                &btypes.Address{Net: 200, Adr: []byte{0x02}},
		}
		e := NewEncoder()
		e.NPDU(&n)
		b := e.Bytes()
		d := NewDecoder(b)
		var out btypes.NPDU
		d.NPDU(&out)
		if !out.IsNetworkLayerMessage {
			t.Error("IsNetworkLayerMessage should be true")
		}
		if !out.ExpectingReply {
			t.Error("ExpectingReply should be true")
		}
	})

	t.Run("hop count with max value", func(t *testing.T) {
		input := []byte{0x01, 0x20, 0x00, 0x64, 0x01, 0x01, 0xFF}
		d := NewDecoder(input)
		n := &btypes.NPDU{}
		d.NPDU(n)
		if n.HopCount != 0xFF {
			t.Errorf("HopCount = %d, expected 0xFF", n.HopCount)
		}
	})

	t.Run("vendor id boundary", func(t *testing.T) {
		n := btypes.NPDU{
			Version:                 1,
			IsNetworkLayerMessage:   true,
			NetworkLayerMessageType: 0x80,
			VendorId:                0xFFFF,
		}
		e := NewEncoder()
		e.NPDU(&n)
		b := e.Bytes()
		d := NewDecoder(b)
		var out btypes.NPDU
		d.NPDU(&out)
		if out.NetworkLayerMessageType != 0x80 {
			t.Errorf("NetworkLayerMessageType = %d, expected 0x80", out.NetworkLayerMessageType)
		}
	})
}
