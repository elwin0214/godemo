package sock

type Message struct {
	Id   uint32
	Body interface{}
}

func NewMessage(id uint32, body interface{}) *Message {
	return &Message{Id: id, Body: body}
}
