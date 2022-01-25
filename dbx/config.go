package dbx

import (
	"fmt"
	"net/url"
)

type Config struct {
	Host        string `valid:"required" schema:"Хост"`
	Port        int    `valid:"required" schema:"Порт"`
	Database    string `valid:"required" schema:"База данных"`
	Username    string `schema:"Логин"`
	Password    string `schema:"Пароль"`
	Schema      string `schema:"Схема"`
	MaxOpenConn int    `schema:"Максимально количество соединений,если <=0 - используется значение по умолчанию равное cpu * 10"`
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
	u.RawQuery = query.Encode()
	return u.String()
}
