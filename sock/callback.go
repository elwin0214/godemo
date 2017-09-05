package sock

import (
	"io"
)

type ConnectionCallBack func(con *Connection)
type ReadCallBack func(con *Connection, msg *Message)
type CodecBuild func(reader io.Reader, writer io.Writer) Codec

type Flusher interface {
	Flush() error
}
