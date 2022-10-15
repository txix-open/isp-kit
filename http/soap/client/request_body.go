package client

import (
	"encoding/xml"
)

type RequestBody interface {
	Body() ([]byte, error)
}

type PlainXml struct {
	Value []byte
}

func (p PlainXml) Body() ([]byte, error) {
	return p.Value, nil
}

type Any struct {
	Value any
}

func (a Any) Body() ([]byte, error) {
	return xml.Marshal(a.Value)
}
