package fitfile

import (
	"os"
	"testing"
)

func Test(t *testing.T) {
	f, err := os.Open("activity.fit")
	if err != nil {
		t.Error(err)
	}
	defer f.Close()

	dec := NewDecoder(f)
	c := &Content{}

	err = dec.Decode(c)
	if err != nil {
		t.Error(err)
	}

	assertEqual(t, c.Header.Size, 12)
	assertEqual(t, c.Header.ProtocolVersion, 16)
	assertEqual(t, c.Header.ProfileVersion, 100)
	assertEqual(t, c.Header.DataSize, uint32(757))
	assertEqual(t, c.Header.DataType, ".FIT")
}

func assertEqual(t *testing.T, a, b interface{}) {
	if a != b {
		t.Error(a, b)
	}
}
