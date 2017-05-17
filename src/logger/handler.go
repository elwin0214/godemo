package logger

import (
	"io"
	"os"
)

type NullHandler struct {
}

func NewNullHandler() *NullHandler {
	return &NullHandler{}
}

func (h *NullHandler) Write(buffer []byte) (n int, err error) {
	return len(buffer), nil
}

func (h *NullHandler) Close() error {
	return nil
}

type StreamHandler struct {
	w io.Writer
}

func NewStreamHandler(w io.Writer) *StreamHandler {
	return &StreamHandler{w: w}
}

func (h *StreamHandler) Write(buffer []byte) (n int, err error) {
	return h.w.Write(buffer)
}

func (h *StreamHandler) Close() error {
	return nil
}

var DefaultHandler Handler = NewStreamHandler(os.Stdout)
