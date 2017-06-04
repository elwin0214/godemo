package sock

import (
	"bufio"
	"encoding/binary"
	"errors"
	"io"
	. "logger"
)

type Codec interface {
	Decode() (interface{}, error)
	Encode(interface{}) error
}

type FixedLenCodec struct {
	byteOrder binary.ByteOrder
	reader    io.Reader
	writer    io.Writer
}

func NewFixedLenCodec(r io.Reader, w io.Writer, isLittle bool) *FixedLenCodec {
	if isLittle {
		return &FixedLenCodec{reader: r, writer: w, byteOrder: binary.LittleEndian}
	} else {
		return &FixedLenCodec{reader: r, writer: w, byteOrder: binary.BigEndian}
	}
}

func FixedLenCodecBuild(reader io.Reader, writer io.Writer) Codec {
	return NewFixedLenCodec(reader, writer, true)
}

func (c *FixedLenCodec) Decode() (interface{}, error) {
	header := make([]byte, 4, 4)
	_, err := io.ReadFull(c.reader, header)
	if err != nil {
		return nil, err
	}
	var length uint32
	length = c.byteOrder.Uint32(header)
	LOG.Info("%d\n", length)
	body := make([]byte, length, length)
	_, err = io.ReadFull(c.reader, body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func (c *FixedLenCodec) Encode(msg interface{}) error {
	body, ok := msg.([]byte)
	if !ok {
		return errors.New("fail cast to []byte")
	}
	var length uint32
	length = uint32(len(body))
	header := make([]byte, 4, 4)
	c.byteOrder.PutUint32(header, length)
	//Write writes len(p) bytes from p to the underlying data stream.
	//It returns the number of bytes written from p (0 <= n <= len(p)) and any error encountered
	// that caused the write to stop early. Write must return a non-nil error if it returns n < len(p).
	// Write must not modify the slice data, even temporarily.
	_, err := c.writer.Write(header)
	if err != nil {
		return err
	}
	_, err = c.writer.Write(body)
	return err
}

type LineCodec struct {
	writer io.Writer
	reader *bufio.Reader
}

func NewLineCodec(reader io.Reader, writer io.Writer) *LineCodec {
	return &LineCodec{writer: writer, reader: bufio.NewReader(reader)}
}

func LineCodecBuild(reader io.Reader, writer io.Writer) Codec {
	return NewLineCodec(reader, writer)
}

func (c *LineCodec) Decode() (interface{}, error) {
	line, err := c.reader.ReadSlice('\n')
	if err != nil {
		return nil, err
	}
	buf := make([]byte, len(line)-1, len(line)-1)
	copy(buf, line[0:len(line)-1])
	return buf, nil
}

func (c *LineCodec) Encode(msg interface{}) error {
	buf, ok := msg.([]byte)
	if !ok {
		return errors.New("fail cast to []byte")
	}
	_, err := c.writer.Write(buf)
	if err != nil {
		return err
	}
	_, err = c.writer.Write([]byte{'\n'})
	return err
}
