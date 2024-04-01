package client

import (
	"encoding/xml"

	"github.com/txix-open/isp-kit/http/soap"
)

type embedEnvelope struct {
	XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Envelope"`

	Header *soap.Header
	Body   embedBody
}

type embedBody struct {
	XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Body"`

	Content []byte      `xml:",innerxml"`
	Fault   *soap.Fault `xml:",omitempty"`
}
