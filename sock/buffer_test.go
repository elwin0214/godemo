package sock

import (
	"bytes"
	"strings"
	"testing"
)

func Test_ReadFrom(t *testing.T) {
	line := "abc"
	line2 := "abc"

	buffer := NewBuffer(6, 16)
	reader := strings.NewReader(line)
	n, _ := buffer.ReadFrom(reader)
	t.Logf("%t\n", line == string(buffer.Bytes()))
	t.Logf("%t\n", line == string([]byte(line2)))

	t.Logf("%t\n", n != 3)

	if n != 3 || line != string(buffer.Bytes()) {
		t.Errorf("line is |%s|, buffer is |%s|, read size is %d\n", line, string(buffer.Bytes()), n)
	}
}

func Test_TooLarge(t *testing.T) {
	buffer := NewBuffer(12, 16)
	for i := 0; i < 4; i++ {
		t.Logf("writable = %d, len = %d\n", buffer.T_writable(), buffer.Len())

		reader := strings.NewReader("abcd")
		n, err := buffer.ReadFrom(reader)
		if n == 0 {
			t.Errorf("read size is %d \n", n)
		}
		if err != nil {
			t.Errorf("error is %s\n", err.Error())
		}
	}
	t.Logf("buffer is %s\n", string(buffer.Bytes()))
	reader := strings.NewReader("abcd")
	_, err := buffer.ReadFrom(reader)

	if err == nil {
		t.Errorf("error should be Too Large buffer")
	}
}
func Test_Truncate(t *testing.T) {
	line := "abc"
	buffer := NewBuffer(6, 16)
	reader := strings.NewReader(line)
	buffer.ReadFrom(reader)
	buffer.Truncate(1)
	if "a" != string(buffer.Bytes()) {
		t.Errorf("buffer is %s\n", string(buffer.Bytes()))
	}
}

func Test_Read(t *testing.T) {
	line := "abc"
	buffer := NewBuffer(6, 16)
	reader := strings.NewReader(line)
	buffer.ReadFrom(reader)
	t.Logf("buf is %s\n", string(buffer.Bytes()))

	a := make([]byte, 2, 2)
	n, err := buffer.Read(a)
	t.Logf("n = %d err = %v\n", n, err)

	t.Logf("buf is %s\n", string(buffer.Bytes()))
	t.Logf("a is %s\n", string(a))

	if "ab" != string(a) {
		t.Errorf("a is %s\n", string(a))
	}
	c := make([]byte, 2, 2)
	buffer.Read(c)
	if 'c' != c[0] {
		t.Errorf("c is %s\n", string(c[0]))
	}
	d := make([]byte, 2, 2)
	_, err = buffer.Read(d)
	if nil == err {
		t.Errorf("can get EOF\n")
	}
}

func Test_Find(t *testing.T) {
	buf := NewBufferString("abc\r\n")
	pos := buf.FindCRLF(0)
	if pos != 3 {
		t.Errorf("pos is %d\n", pos)
	}

}

func Test_NoFind(t *testing.T) {
	buf := NewBufferString("abc\r")
	pos := buf.FindCRLF(0)
	if pos != -1 {
		t.Errorf("pos is %d\n", pos)
	}
}

func Test_Find_Pos(t *testing.T) {
	buf := NewBufferString("set a 0 1 1\r\na\r\n")
	pos := buf.FindCRLF(0)
	if pos != 11 {
		t.Errorf("pos is %d\n", pos)
	}
}

func Test_Skip(t *testing.T) {
	buf := NewBufferString("abc")
	buf.Skip(2)
	if "c" != string(buf.Bytes()) {
		t.Errorf("buf is '%s'\n", string(buf.Bytes()))
	}
	buf.Skip(1)
	if "" != string(buf.Bytes()) {
		t.Errorf("buf is '%s'\n", string(buf.Bytes()))
	}
}

func Test_Compact(t *testing.T) {
	reader := bytes.NewBufferString("123456789")
	buf := NewBuffer(16, -1)
	buf.ReadFrom(reader)
	b := make([]byte, 6, 6)
	buf.Read(b)
	//buf.T_Compact()
	t.Logf("b is %s\n", string(b))
	if "789" != string(buf.Bytes()) {
		t.Errorf("buf is [%s]\n", string(buf.Bytes()))
	}
	t.Logf("%s\n", buf.Tostring())

	buf.T_Compact()
	t.Logf("%s\n", buf.Tostring())
	if "789" != string(buf.Bytes()) {
		t.Errorf("buf is [%s]\n", string(buf.Bytes()))
	}
}

func Test_ReadLarge(t *testing.T) {
	size := 1024 * 1024
	reader := bytes.NewBuffer(make([]byte, size, size))
	t.Logf("reader.len = %d reader.cap = %d\n", reader.Len(), reader.Cap())

	buf := NewBuffer(16, -1)
	//buf := NewBufferString("abc")
	for {
		_, err := buf.ReadFrom(reader)
		t.Logf("buf.len = %d buf.cap = %d\n", buf.Len(), buf.Cap())

		if nil != err {
			t.Logf("error = %s\n", err.Error())
			break
		}
	}
	t.Logf("buf.Len == size is %t\n", (buf.Len() == size))

	t.Logf("buf.len = %d buf.cap = %d\n", buf.Len(), buf.Cap())
}
