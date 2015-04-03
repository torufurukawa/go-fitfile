package fitfile

// Content represents a structured content that is decoded from an FIT File.
type Content struct {
	Header *Header
}

type Header struct {
	Size            int
	ProtocolVersion int
	ProfileVersion  int
	DataSize        uint32
	DataType        string
}
