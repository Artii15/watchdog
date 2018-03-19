package loggers

type Message struct {
	Content []interface{}
	Prefix string
}

func (message *Message) New(prefix string, content ...interface{}) Message {
	return Message{Content: content, Prefix: prefix}
}
