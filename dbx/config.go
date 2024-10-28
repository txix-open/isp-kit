package dbx

import (
	"fmt"
	"net/url"
)

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

func (c Config) Dsn() string {
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
	for key, value := range c.Params {
		query.Set(key, value)
	}
	u.RawQuery = query.Encode()
	return u.String()
}
