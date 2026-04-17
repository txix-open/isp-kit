package publisher

import (
	"github.com/go-stomp/stomp/v3"
)

// Message represents a message to be published to a STOMP broker.
type Message struct {
	// ContentType is the MIME type of the message body.
	ContentType string
	// Body is the raw message content.
	Body []byte
	// Opts are publication options (headers, etc.).
	Opts []PublishOption
}

// Json creates a new message with JSON content type.
func Json(body []byte) *Message {
	return &Message{
		ContentType: "application/json",
		Body:        body,
		Opts:        nil,
	}
}

// PlainText creates a new message with plain text content type.
func PlainText(body []byte) *Message {
	return &Message{
		ContentType: "plain/text",
		Body:        body,
		Opts:        nil,
	}
}

// WithHeader adds a header to the message.
func (m *Message) WithHeader(key string, value string) *Message {
	m.Opts = append(m.Opts, stomp.SendOpt.Header(key, value))
	return m
}
