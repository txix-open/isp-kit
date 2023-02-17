package publisher

import (
	"github.com/go-stomp/stomp/v3"
)

type Message struct {
	ContentType string
	Body        []byte
	Opts        []PublishOption
}

func Json(body []byte) *Message {
	return &Message{
		ContentType: "application/json",
		Body:        body,
		Opts:        nil,
	}
}

func PlainText(body []byte) *Message {
	return &Message{
		ContentType: "plain/text",
		Body:        body,
		Opts:        nil,
	}
}

func (m *Message) WithHeader(key string, value string) *Message {
	m.Opts = append(m.Opts, stomp.SendOpt.Header(key, value))
	return m
}
