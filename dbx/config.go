package dbx

import (
	"fmt"
	"net/url"
)

// Config holds the configuration for establishing a database connection.
type Config struct {
	Host        string            `validate:"required" schema:"Хост"`
	Port        int               `validate:"required" schema:"Порт"`
	Database    string            `validate:"required" schema:"База данных"`
	Username    string            `schema:"Логин"`
	Password    string            `schema:"Пароль"`
	Schema      string            `schema:"Схема"`
	MaxOpenConn int               `schema:"Максимально количество соединений,если <=0 - используется значение по умолчанию равное cpu * 10"`
	Params      map[string]string `schema:"Дополнительные параметры подключения"`
}

// Dsn generates a PostgreSQL connection string from the configuration.
// The applicationName parameter is included as a connection parameter.
func (c Config) Dsn(applicationName string) string {
	u := url.URL{
		Scheme: "postgresql",
		User:   nil,
		Host:   fmt.Sprintf("%s:%d", c.Host, c.Port),
		Path:   c.Database,
	}
	if c.Username != "" {
		u.User = url.UserPassword(c.Username, c.Password)
	}
	query := url.Values{}
	if c.Schema != "" {
		query.Set("search_path", c.Schema)
	}

	if applicationName != "" {
		query.Set("application_name", applicationName)
	}

	for key, value := range c.Params {
		query.Set(key, value)
	}
	u.RawQuery = query.Encode()
	return u.String()
}
