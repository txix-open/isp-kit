package kafkax

import (
	"crypto/tls"
	"encoding/base64"
	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go/sasl"
	"github.com/segmentio/kafka-go/sasl/plain"
	"github.com/segmentio/kafka-go/sasl/scram"
)

const (
	AuthTypePlain   = "SASL/PLAIN"
	AuthTypeSCRAM   = "SASL/SCRAM"
	ScramTypeSHA256 = "SHA256"
	ScramTypeSHA512 = "SHA512"
)

type Auth struct {
	Mechanism *string `validate:"oneof=SASL/PLAIN SASL/SCRAM" schema:"Механизм аутентификации может принимать значения 'SASL/PLAIN' или 'SASL/SCRAM'"`
	ScramType *string `validate:"oneof=SHA256 SHA512" schema:"Алгоритм используемый в механизме SASL/SCRAM, может принимать значения 'SHA256' или 'SHA512'"`
	Username  string  `validate:"required" schema:"Логин"`
	Password  string  `validate:"required" schema:"Пароль"`
}

type TLS struct {
	Certificate string `schema:"Сертификат для подключения"`
	PrivateKey  string `schema:"Закрытый ключ"`
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

func ScramAuth(auth *Auth) (sasl.Mechanism, error) {
	if auth == nil {
		return nil, nil
	}

	if auth.ScramType == nil {
		return nil, errors.Errorf("scramType is required")
	}

	var algorithm scram.Algorithm
	switch *auth.ScramType {
	case ScramTypeSHA256:
		algorithm = scram.SHA256
	case ScramTypeSHA512:
		algorithm = scram.SHA512
	default:
		return nil, errors.Errorf("unexpected scram type %s", *auth.ScramType)
	}

	mechanism, err := scram.Mechanism(algorithm, auth.Username, auth.Password)
	if err != nil {
		return nil, errors.WithMessage(err, "get auth mechanism")
	}

	return mechanism, nil
}

func setupSASLMechanism(mechanismType string, auth *Auth) (sasl.Mechanism, error) {
	var saslMechanism sasl.Mechanism

	switch mechanismType {
	case AuthTypePlain:
		saslMechanism = PlainAuth(auth)
	case AuthTypeSCRAM:
		var err error
		saslMechanism, err = ScramAuth(auth)
		if err != nil {
			return nil, errors.WithMessage(err, "get scram auth mechanism")
		}
	default:
		return nil, errors.Errorf("unexpected sasl mechanism: %s", mechanismType)
	}

	return saslMechanism, nil
}

func setupTLS(cfg *TLS) (*tls.Config, error) {
	if cfg == nil {
		return nil, nil
	}

	rawCert, err := base64.StdEncoding.DecodeString(cfg.Certificate)
	if err != nil {
		return nil, errors.WithMessage(err, "decode base64 tls certificate")
	}

	return &tls.Config{
		Certificates: []tls.Certificate{
			{
				Certificate: [][]byte{rawCert},
				PrivateKey:  cfg.PrivateKey,
			},
		},
	}, nil
}
