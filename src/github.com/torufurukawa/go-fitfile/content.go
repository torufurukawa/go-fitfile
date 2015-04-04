package fitfile

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// Content represents a structured content that is decoded from an FIT File.
type Content struct {
	Header       *Header
	HeaderCRC    uint16
	DataMessages []DataMessage
}

// NewContent constructs a content.
func NewContent() *Content {
	return &Content{
		Header: &Header{},
	}
}

// Header contains FIT File header parameters except CRC.
type Header struct {
	Size            uint8
	ProtocolVersion uint8
	ProfileVersion  uint16
	DataSize        uint32
	DataType        [4]byte
}

// DataMessage contains data fields.
type DataMessage struct {
	Type         LocalMessageType
	SerialNumber *DataField
}

// DataField contains Value and Unit related to field.
type DataField struct {
	Value interface{}
	Unit  string
}

// NewDataMassage constructs from buf based on def and return it.
func NewDataMessage(def *Definition, t LocalMessageType, buf []byte) (*DataMessage, error) {
	m := &DataMessage{Type: t}
	reader := bytes.NewReader(buf)
	_ = reader

	for _, f := range def.Fields {
		fmt.Println(f.Number, f.Size, f.BaseType)

		// read
		b := make([]byte, f.Size)
		_ = b

		baseTypeNumber := f.BaseType & 0x0F
		fmt.Println("base type number:", baseTypeNumber)
		switch baseTypeNumber {
		// TODO: implement all base type numbers
		case 12:
			n, err := reader.Read(b)
			if n < len(b) || err != nil {
				return nil, fmt.Errorf("cannot decode datamessage: %d %s", n, err)
			}
			switch t {
			case 0:
				switch f.Number {
				case 3:
					var v uint32
					binary.Read(bytes.NewReader(b), def.ByteOrder, &v)
					m.SerialNumber = &DataField{v, ""}
				}
			}
			fmt.Println("type", t)
			fmt.Println("number: ", f.Number)

		default:
			return nil, nil
		}
	}

	m.SerialNumber = &DataField{0, ""}
	return m, nil
}
