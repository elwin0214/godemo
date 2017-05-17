package demo

import (
	"testing"
)

func Test_Copy(t *testing.T) {
	s := make([]byte, 8, 8)
	s[0] = '0'
	s[1] = '1'
	s[2] = '2'
	s[3] = '3'
	s[4] = '4'
	s[5] = '5'
	s[6] = '6'
	s[7] = '7'

	copy(s, s[2:5])
	t.Logf("%s\n", string(s))
}

func Test_Slice(t *testing.T) {
	str := "abc"
	//slice := str[0:2]
	//slice[0] = '1' //error
	slice2 := []byte(str)
	slice2[0] = '1'

	t.Logf("len = %d cap = %d \n", len(slice2), cap(slice2))

	if str != "abc" {
		t.Errorf("string can be modified\n")
	}

	slice3 := make([]byte, 1, 1)
	slice3[0] = 'a'
	s3 := string(slice3)

	t.Logf("%s\n", s3)
	slice3[0] = 'b'
	t.Logf("%s\n", s3)

	if s3 != "a" {
		t.Errorf("string can be modified\n")
	}

	s4 := make([]byte, 6)
	t.Logf("s4.len = %d\n", len(s4))

	s4[0] = 'a'
	s4[1] = 'b'
	s4[2] = 'c'
	t.Logf("s4 = %s s4[0:2] = %s\n", string(s4), string(s4[0:2]))
	s5 := s4[1:2]
	t.Logf("s5 = %s s5[0:1] = %s\n", string(s5), string(s5[0:1]))

	t.Logf("len(s4[1:]) = %d\n", len(s4[1:]))
}
