package sock

import (
	"bufio"
	"bytes"
	. "sock"
	"testing"
)

func Test_ReadSlice(t *testing.T) {

	buffer := bytes.NewBuffer([]byte("abc\n"))
	reader := bufio.NewReader(buffer)
	buf, err := reader.ReadSlice('\n')
	t.Logf("%s %v\n", string(buf), err)
	buf, err = reader.ReadSlice('\n')
	t.Logf("%s %s\n", string(buf), err)
}

func Test_FixedLengthCodec(t *testing.T) {
	writeMsg := "abc"
	buf := make([]byte, 0, 12)
	buffer := bytes.NewBuffer(buf)
	codec := FixedLenCodecBuild(buffer, buffer)
	codec.Encode([]byte(writeMsg))
	body, _ := codec.Decode()
	readMsg, _ := body.([]byte)
	if writeMsg != string(readMsg) {
		t.Errorf("result is '%s'\n", string(readMsg))
	}
}

func Test_LineCodec(t *testing.T) {
	writeMsg := "abc"
	buf := make([]byte, 0, 12)
	buffer := bytes.NewBuffer(buf)
	codec := LineCodecBuild(buffer, buffer)
	codec.Encode([]byte(writeMsg))
	body, _ := codec.Decode()
	readMsg, _ := body.([]byte)
	if writeMsg != string(readMsg) {
		t.Errorf("result is '%s'\n", string(readMsg))
	}
}
