package sock

import (
	"errors"
	"fmt"
	"io"
)

type Buffer struct {
	data   []byte
	off    int
	maxCap int
}

func NewBuffer(n int, max int) *Buffer {
	return &Buffer{
		data:   make([]byte, 0, n),
		off:    0,
		maxCap: max,
	}
}

func NewBufferString(s string) *Buffer {
	//if len(s) < = 0
	data := []byte(s)
	buffer := &Buffer{
		data:   data,
		off:    0,
		maxCap: -1,
	}
	return buffer
}

var ErrTooLarge = errors.New("Buffer: too large")

const minWritableRoom = 16

func (b *Buffer) ReadFrom(reader io.Reader) (int, error) {
	if b.readable() <= 0 {
		b.Truncate(0)
	}

	if b.writable() < minWritableRoom {
		if minWritableRoom <= b.off {
			b.Compact()
		} else {
			b.grow()
		}
	}
	if b.writable() <= 0 {
		return 0, errors.New(fmt.Sprintf("buffer too large{len = %d cap = %d off = %d maxcap = %d}\n", len(b.data), cap(b.data), b.off, b.maxCap))
	}
	n, err := reader.Read(b.data[len(b.data):cap(b.data)])
	b.data = b.data[0 : len(b.data)+n]
	return n, err
}

func (b *Buffer) Read(buf []byte) (int, error) {
	size := len(buf)
	if size == 0 {
		return 0, nil
	}
	if size > b.Len() {
		size = b.Len()
	}
	if size == 0 {
		return 0, io.EOF
	}
	copy(buf, b.data[b.off:b.off+size])
	b.off = b.off + size
	return size, nil
}

func (b *Buffer) Truncate(n int) {
	if n < 0 || n > len(b.data) {
		panic("truncation out of range")
	}
	if 0 == n {
		b.off = 0
	}
	b.data = b.data[0 : b.off+n]

}

func (b *Buffer) Compact() {
	size := len(b.data) - b.off
	copy(b.data[0:size], b.data[b.off:len(b.data)])
	b.off = 0
	b.data = b.data[0:size]
}

func (b *Buffer) T_Compact() {
	b.Compact()
}

func (b *Buffer) Bytes() []byte {
	return b.data[b.off:]
}

func (b *Buffer) Len() int {
	return len(b.data) - b.off
}

func (b *Buffer) Cap() int {
	return cap(b.data)
}
func (b *Buffer) readable() int {
	return len(b.data) - b.off
}

func (b *Buffer) writable() int {
	return cap(b.data) - len(b.data)
}

func (b *Buffer) T_writable() int {
	return b.writable()
}

func (b *Buffer) Skip(size int) {
	if b.off+size > len(b.data) {
		panic("over range of data")
	}
	b.off = b.off + size
}

func (b *Buffer) FindCRLF(from int) int {
	if from < 0 {
		from = 0
	}
	slice := b.Bytes()
	for i := from; i+1 < len(slice); i++ {
		if slice[i] == '\r' && slice[i+1] == '\n' {
			return i
		}
	}
	return -1
}
func (b *Buffer) grow() bool {
	if b.maxCap > 0 && cap(b.data) >= b.maxCap {
		return false
	}
	targetCap := 2 * cap(b.data)
	if targetCap > b.maxCap && b.maxCap > 0 {
		targetCap = b.maxCap
	}

	buf := make([]byte, targetCap)
	if len(b.data)-b.off > 0 {
		copy(buf, b.data[b.off:len(b.data)])
		b.data = buf[0 : len(b.data)-b.off]
		b.off = 0
	} else {
		b.data = buf[0:0]
		b.off = 0
	}

	return true
}
func (b *Buffer) T_grow() bool {
	return b.grow()
}

func (b *Buffer) Tostring() string {
	return fmt.Sprintf("{off=%d,len=%d,cap=%d,data=%s}", b.off, len(b.data), cap(b.data), string(b.Bytes()))
}
