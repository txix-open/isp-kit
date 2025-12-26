// nolint:ireturn,lll,unparam,gosec
package kafkax

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/pkg/errors"
	"github.com/twmb/franz-go/pkg/sasl"
	"github.com/twmb/franz-go/pkg/sasl/plain"
	"github.com/twmb/franz-go/pkg/sasl/scram"
	"strings"
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
	RootCA             *string `schema:"Корневой сертификат"`
	PrivateKey         *string `schema:"Закрытый ключ"`
	Certificate        *string `schema:"Сертификат клиента"`
	InsecureSkipVerify bool    `schema:"Пропуск проверки сертификатов"`
}

func PlainAuth(auth *Auth) sasl.Mechanism {
	if auth == nil {
		return nil
	}

	return plain.Auth{
		User: auth.Username,
		Pass: auth.Password,
	}.AsMechanism()
}

func ScramAuth(auth *Auth) (sasl.Mechanism, error) {
	if auth == nil {
		return nil, nil // nolint:nilnil
	}

	if auth.ScramType == nil {
		return nil, errors.Errorf("scramType is required")
	}

	scramAuth := scram.Auth{
		User: auth.Username,
		Pass: auth.Password,
	}

	switch strings.ToUpper(*auth.ScramType) {
	case ScramTypeSHA256:
		return scramAuth.AsSha256Mechanism(), nil
	case ScramTypeSHA512:
		return scramAuth.AsSha512Mechanism(), nil
	default:
		return nil, errors.Errorf("unexpected scram type %s", *auth.ScramType)
	}
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
		return nil, nil // nolint:nilnil
	}

	caCertPool := x509.NewCertPool()
	if cfg.RootCA != nil {
		caCertPool.AppendCertsFromPEM([]byte(*cfg.RootCA))
	}

	var certificates []tls.Certificate
	if cfg.Certificate != nil && cfg.PrivateKey != nil {
		cert := tls.Certificate{
			Certificate: [][]byte{[]byte(*cfg.Certificate)},
			PrivateKey:  *cfg.PrivateKey,
		}
		certificates = append(certificates, cert)
	}

	return &tls.Config{
		InsecureSkipVerify: cfg.InsecureSkipVerify,
		RootCAs:            caCertPool,
		Certificates:       certificates,
	}, nil
}
