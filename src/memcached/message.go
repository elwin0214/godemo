package memcached

type ParserError struct {
	text string
}

func (pe *ParserError) Error() string {
	return pe.text
}

func newParserError(text string) *ParserError {
	return &ParserError{text: text}
}

type MemRequest struct {
	Op      Code
	Key     string
	Flags   uint32
	Exptime uint32
	Bytes   uint16
	Data    []byte
	Value   uint32 // for counter
	Err     string
}

type MemResponse struct {
	Op      Code
	Result  bool
	Key     string
	Flags   uint32
	Exptime uint32
	Bytes   uint16
	Data    []byte
	Value   uint32
	Err     string
}
