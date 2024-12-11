package kafkax

import (
	"github.com/segmentio/kafka-go/sasl"
	"github.com/segmentio/kafka-go/sasl/plain"
)

type Auth struct {
	Username string `validate:"required" schema:"Логин"`
	Password string `validate:"required" schema:"Пароль"`
}

func PlainAuth(auth *Auth) sasl.Mechanism {
	if auth == nil {
		return nil
	}

	return plain.Mechanism{
		Username: auth.Username,
		Password: auth.Password,
	}
}
