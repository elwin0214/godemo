package memcached

import (
	"testing"
)

func Test_Simple(t *testing.T) {
	data := []byte("a b c")
	sep := byte(' ')
	tk := NewTokenizer(data, sep)

	var words = [3]string{"a", "b", "c"}

	for i := 0; i < len(words); i++ {
		ok, r := tk.Next()
		if !ok || string(r) != words[i] {
			t.Errorf("word is %s\n", string(r))
		}
	}

	ok, r := tk.Next()
	if ok {
		t.Errorf("word is %s\n", string(r))
	}
}

func Test_Blank(t *testing.T) {
	data := []byte(" a b c ")
	sep := byte(' ')
	tk := NewTokenizer(data, sep)
	var words = [3]string{"a", "b", "c"}
	for i := 0; i < len(words); i++ {
		ok, r := tk.Next()
		if !ok || string(r) != words[i] {
			t.Errorf("word is %s\n", string(r))
		}
	}
	ok, r := tk.Next()
	if ok {
		t.Errorf("word is %s\n", string(r))
	}
}

func Test_MultiChar(t *testing.T) {
	data := []byte(" aaa bbb  ccc ")
	sep := byte(' ')
	tk := NewTokenizer(data, sep)
	var words = [3]string{"aaa", "bbb", "ccc"}
	for i := 0; i < len(words); i++ {
		ok, r := tk.Next()
		if !ok || string(r) != words[i] {
			t.Errorf("word is %s\n", string(r))
		}
	}
	ok, r := tk.Next()
	if ok {
		t.Errorf("word is %s\n", string(r))
	}
}

func Test_Next(t *testing.T) {
	data := []byte("abc")
	sep := byte(' ')
	tk := NewTokenizer(data, sep)

	ok, r := tk.Next()
	if !ok || string(r) != "abc" {
		t.Errorf("word is %s\n", string(r))
	}
	if ok, r = tk.Next(); ok {
		t.Errorf("next word exist\n")
	}
}
