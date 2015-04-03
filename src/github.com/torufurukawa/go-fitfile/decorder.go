package fitfile

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

const MINIMAM_HEADER_SIZE = 12

// A Decoder reads and decodes FIT File records from an input stream.
type Decoder struct {
	reader io.Reader
}

// NewDecoder returns a new decoder that reads from r.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r}
}

// Decord parses the FIT File records from the stream and stores the result in
// the value pointed to by c.
func (dec *Decoder) Decode(c *Content) error {
	if err := dec.decodeHeader(c); err != nil {
		return err
	}
	return nil
}

func (dec *Decoder) decodeHeader(c *Content) error {
	header := &Header{}
	buf := make([]byte, MINIMAM_HEADER_SIZE)

	// read
	n, err := dec.reader.Read(buf)
	if n < MINIMAM_HEADER_SIZE {
		return fmt.Errorf("header is too small %d.", n)
	}
	if err != nil {
		return err
	}

	header.Size = int(buf[0])
	header.ProtocolVersion = int(buf[1])

	var profileVersion uint16
	if err = extractAt(buf, 2, &profileVersion); err != nil {
		return fmt.Errorf("cannot extract profile version header: %s", err)
	}
	header.ProfileVersion = int(profileVersion)

	var dataSize uint32
	if err = extractAt(buf, 4, &dataSize); err != nil {
		return fmt.Errorf("cannot extract data size header: %s", err)
	}
	header.DataSize = dataSize

	header.DataType = string(buf[8 : 8+4])
	if header.DataType != ".FIT" {
		return fmt.Errorf("invalid data type header:", header.DataType)
	}

	// store
	c.Header = header
	return nil
}

func extractAt(data []byte, pos int, v interface{}) error {
	buf := bytes.NewBuffer(data[pos : pos+binary.Size(v)])
	err := binary.Read(buf, binary.LittleEndian, v)
	return err
}
