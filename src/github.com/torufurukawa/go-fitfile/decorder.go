package fitfile

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

const (
	MINIMAM_HEADER_SIZE          = 12
	HEADER_CRC_SIZE              = 2
	SIGNATURE                    = ".FIT"
	MINIMAL_RECORD_SIZE          = 2
	DEF_MESSAGE_FIXED_BYTE_COUNT = 5
	BYTE_COUNT_PER_FIELD         = 3
)

type RecordHeader struct {
	MessageType      MessageType
	HeaderType       HeaderType
	LocalMessageType LocalMessageType
}

func (h *RecordHeader) IsDefinition() bool {
	return h.MessageType == MessageDefinition
}

type MessageType int

const (
	MessageDefinition MessageType = 0
	MessageData       MessageType = 1
)

type HeaderType uint8

const (
	NormalHeader     HeaderType = 0
	CompressedHeader            = 1
)

type LocalMessageType int

// A Decoder reads and decodes FIT File records from an input stream.
type Decoder struct {
	reader      *bufio.Reader
	definitions map[LocalMessageType]Definition
}

// NewDecoder returns a new decoder that reads from r.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		reader:      bufio.NewReader(r),
		definitions: make(map[LocalMessageType]Definition),
	}
}

// Decord parses the FIT File records from the stream and stores the result in
// the value pointed to by c.
func (dec *Decoder) Decode(c *Content) error {
	if err := dec.decodeHeader(c); err != nil {
		return err
	}

	if dec.needsHeaderCRCValidation(c) {
		if err := dec.decodeHeaderCRC(c); err != nil {
			return err
		}
		if err := dec.validateHeaderCRC(c); err != nil {
			return err
		}
	}

	if err := dec.decodeDataRecords(c); err != nil {
		return err
	}

	return nil
}

func (dec *Decoder) decodeHeader(c *Content) error {
	buf := make([]byte, MINIMAM_HEADER_SIZE)

	// load to buffer
	n, err := dec.reader.Read(buf)
	if n < MINIMAM_HEADER_SIZE {
		return fmt.Errorf("header is too small %d.", n)
	}
	if err != nil {
		return err
	}

	// read header
	bufReader := bytes.NewReader(buf)
	err = binary.Read(bufReader, binary.LittleEndian, c.Header)
	if err != nil {
		return fmt.Errorf("failed to read header: %s", err)
	}

	// validate data type signature
	if string(c.Header.DataType[:len(SIGNATURE)]) != SIGNATURE {
		return fmt.Errorf("invalid data type:", c.Header.DataType)
	}

	return nil
}

func (dec *Decoder) needsHeaderCRCValidation(c *Content) bool {
	return c.Header.Size == uint8(MINIMAM_HEADER_SIZE+HEADER_CRC_SIZE)
}

func (dec *Decoder) decodeHeaderCRC(c *Content) error {
	// prepare buffer
	buf := make([]byte, HEADER_CRC_SIZE)
	n, err := dec.reader.Read(buf)
	if n < HEADER_CRC_SIZE {
		return fmt.Errorf("cannot find header CRC: %d", n)
	}
	if err != nil {
		return fmt.Errorf("failed to read header CRC: %s", err)
	}

	// read
	err = binary.Read(bytes.NewReader(buf), binary.LittleEndian, &c.HeaderCRC)
	if err != nil {
		return fmt.Errorf("failed to read header CRC: %s", err)
	}

	return nil
}

func (dec *Decoder) validateHeaderCRC(c *Content) error {
	// always valid if header CRC is zero
	if c.HeaderCRC == uint16(0) {
		return nil
	}

	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.LittleEndian, c.Header); err != nil {
		return fmt.Errorf("failed to encode header: %s", err)
	}

	// compute CRC
	var crc uint16
	for _, val := range buf.Bytes() {
		crc = calcCRC(crc, val)
	}

	// error if CRCs unmached
	if crc != c.HeaderCRC {
		return fmt.Errorf("header CRC should be %x, but %x", c.HeaderCRC, crc)
	}

	return nil
}

func (dec *Decoder) decodeDataRecords(c *Content) error {
	for {
		h := &RecordHeader{}
		if err := dec.decodeRecordHeader(h); err != nil {
			return err
		}

		if h.IsDefinition() {
			if err := dec.decodeDefinitionMessage(h); err != nil {
				return err
			}
		} else {
			// decode data massage
			if err := dec.decodeDataMessage(h, c); err != nil {
				return err
			}

			// TODO: remove this unless failed
			return nil
		}
	}

	return nil
}

func (dec *Decoder) decodeRecordHeader(h *RecordHeader) error {
	// read header byte
	buf := make([]byte, 1)
	n, err := dec.reader.Read(buf)
	if n < len(buf) || err != nil {
		fmt.Errorf("cannot read record header: %d %s", n, err)
	}
	b := buf[0]

	// parse
	if b&0x80 == 0x80 {
		h.HeaderType = NormalHeader
	} else {
		h.HeaderType = CompressedHeader
	}

	if b&0x40 == 0x40 {
		h.MessageType = MessageDefinition
	} else {
		h.MessageType = MessageData
	}

	h.LocalMessageType = LocalMessageType(int(b & 0x07))

	return nil
}

func (dec *Decoder) decodeDefinitionMessage(h *RecordHeader) error {
	// read fixed content
	buf := make([]byte, DEF_MESSAGE_FIXED_BYTE_COUNT)
	n, err := dec.reader.Read(buf)
	if n < len(buf) || err != nil {
		fmt.Errorf("unexpected EOF")
	}

	def := NewDefinition(buf)

	// read variable contents
	buf = make([]byte, BYTE_COUNT_PER_FIELD*len(def.Fields))
	n, err = dec.reader.Read(buf)
	if n < len(buf) || err != nil {
		fmt.Errorf("unexpected EOF")
	}

	// decode
	for i := 0; i < int(len(def.Fields)); i++ {
		fieldDefinition := FieldDefinition{
			Number:   uint8(buf[i*BYTE_COUNT_PER_FIELD]),
			Size:     uint8(buf[i*BYTE_COUNT_PER_FIELD+1]),
			BaseType: buf[i*BYTE_COUNT_PER_FIELD+2],
		}
		def.Fields[i] = fieldDefinition
	}

	// store definition
	dec.definitions[h.LocalMessageType] = *def

	return nil
}

func (dec *Decoder) decodeDataMessage(h *RecordHeader, c *Content) error {
	// determine bytes to read
	// TODO: should store *Definition instead of Definition
	def := dec.definitions[h.LocalMessageType]
	byteCount := 0
	for _, v := range def.Fields {
		byteCount += int(v.Size)
	}

	// read
	buf := make([]byte, byteCount)
	n, err := dec.reader.Read(buf)
	if n < len(buf) || err != nil {
		return fmt.Errorf("failed to read data message %d %s", n, err)
	}
	fmt.Println(buf)

	// TODO: decode and store
	m, err := NewDataMessage(&def, h.LocalMessageType, buf)
	if err != nil {
		return fmt.Errorf("failed to decode data message %s", err)
	}

	c.DataMessages = append(c.DataMessages, *m)

	return nil
}

type Definition struct {
	GlobalMessageNumber uint16
	Fields              []FieldDefinition
	ByteOrder           binary.ByteOrder
}

func NewDefinition(buf []byte) *Definition {
	var byteOrder binary.ByteOrder
	if buf[1] == 0 {
		byteOrder = binary.BigEndian
	} else {
		byteOrder = binary.LittleEndian
	}

	var globalMessageNumber uint16
	binary.Read(bytes.NewReader(buf[2:3]), byteOrder, globalMessageNumber)

	fieldCount := uint8(buf[4])

	d := &Definition{
		GlobalMessageNumber: globalMessageNumber,
		Fields:              make([]FieldDefinition, fieldCount),
		ByteOrder:           byteOrder,
	}

	return d
}

type FieldDefinition struct {
	Number   uint8
	Size     uint8
	BaseType byte
}
