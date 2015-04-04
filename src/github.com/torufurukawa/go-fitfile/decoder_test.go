package fitfile

import (
	"os"
	"testing"
)

func TestActivity(t *testing.T) {
	f, err := os.Open("Activity.fit")
	if err != nil {
		t.Error(err)
	}
	defer f.Close()

	c := NewContent()
	dec := NewDecoder(f)
	if err = dec.Decode(c); err != nil {
		t.Error(err)
	}

	// header
	assertEqual(t, c.Header.Size, uint8(12))
	assertEqual(t, c.Header.ProtocolVersion, uint8(16))
	assertEqual(t, c.Header.ProfileVersion, uint16(100))
	assertEqual(t, c.Header.DataSize, uint32(757))
	assertEqual(t, string(c.Header.DataType[:4]), ".FIT")

	// definition
	assertEqual(t, dec.definitions[0].GlobalMessageNumber, uint16(0))
	assertEqual(t, dec.definitions[0].Fields[0].Number, uint8(3))
	assertEqual(t, dec.definitions[0].Fields[0].Size, uint8(4))
	assertEqual(t, dec.definitions[0].Fields[0].BaseType, byte(0x8c))

	// data
	assertNotEqual(t, c.DataMessages[0].SerialNumber, nil)
	assertEqual(t, c.DataMessages[0].SerialNumber.Value, "2147483647")
	assertEqual(t, c.DataMessages[0].SerialNumber.Unit, "")
	// TOOD: more tests
}

// TestHeaderCRC is to test only header CRC.
func TestHeaderCRC(t *testing.T) {
	f, err := os.Open("MonitoringFile.fit")
	if err != nil {
		t.Error(err)
	}
	defer f.Close()

	c := NewContent()
	dec := NewDecoder(f)
	if err = dec.Decode(c); err != nil {
		t.Error(err)
	}

	assertEqual(t, c.HeaderCRC, uint16(0xd465))
}

func TestRealActivity(t *testing.T) {
	f, err := os.Open("MonitoringFile.fit")
	if err != nil {
		t.Error(err)
	}
	defer f.Close()

	c := NewContent()
	dec := NewDecoder(f)
	if err = dec.Decode(c); err != nil {
		t.Error(err)
	}

	// TODO: test next line

}

func assertEqual(t *testing.T, a, b interface{}) {
	if a != b {
		t.Error(a, b)
	}
}

func assertNotEqual(t *testing.T, a, b interface{}) {
	if a == b {
		t.Error(a, b)
	}
}
