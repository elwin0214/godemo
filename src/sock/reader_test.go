package sock

import (
	"strings"
	"testing"
)

func Test_ReapeatReadEOF(t *testing.T) {
	reader := strings.NewReader("abc")
	buf := make([]byte, 3, 3)
	n, err := reader.Read(buf)
	t.Log("n = %d, err = %v\n", n, err) // 3 nil
	buf2 := make([]byte, 3, 3)
	n, err = reader.Read(buf2)
	t.Log("n = %d, err = %v\n", n, err) // 0 EOF

	n, err = reader.Read(buf2)
	t.Log("n = %d, err = %v\n", n, err) //

}

func Test_ReaderEOF(t *testing.T) {
	reader := strings.NewReader("abc")
	buf := make([]byte, 4, 4)
	n, err := reader.Read(buf)
	t.Log("n = %d, err = %v\n", n, err) //nil
	buf2 := make([]byte, 3, 3)
	n, err = reader.Read(buf2)
	t.Log("n = %d, err = %v\n", n, err) // EOF

}

func Test_Reader_Part(t *testing.T) {
	reader := strings.NewReader("abc")
	buf := make([]byte, 2, 4)
	n, err := reader.Read(buf)
	t.Log("n = %d, err = %v\n", n, err)

}
