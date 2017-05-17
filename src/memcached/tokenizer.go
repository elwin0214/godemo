package memcached

type Tokenizer struct {
	data []byte
	pos  int32
	sep  byte
}

func NewTokenizer(data []byte, sep byte) *Tokenizer {
	return &Tokenizer{data: data, sep: sep, pos: 0}
}

func (t *Tokenizer) hasNext() bool {
	return t.pos >= 0
}

func (t *Tokenizer) Next() (bool, []byte) {
	if t.pos < 0 {
		return false, nil
	}
	match := false
	start := int32(-1)
	end := int32(-1)
	for i := t.pos; i < int32(len(t.data)); i++ {
		if t.data[i] == t.sep {
			if !match {
				continue
			} else {
				end = i
				break
			}
		} else {
			if !match {
				match = true
				start = i
			}
		}
	}

	if start >= 0 && end < 0 {
		end = int32(len(t.data))
	}
	if !match {
		t.pos = -1
		return false, nil
	} else {
		t.pos = end
		return true, t.data[start:end]
	}
}
