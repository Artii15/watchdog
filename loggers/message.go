package loggers

type Message struct {
	Content string
	Prefix string
}

func (message *Message) New(content, prefix string) Message {
	return Message{Content: content, Prefix: prefix}
}
