package client

import (
	"encoding/xml"

	"github.com/txix-open/isp-kit/http/soap"
)

// embedEnvelope is an internal type for handling raw XML inside SOAP envelopes.
// It is used for PlainXml requests and responses.
type embedEnvelope struct {
	XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Envelope"`

	Header *soap.Header
	Body   embedBody
}

// embedBody represents the body of a SOAP envelope with raw XML content.
// It supports both inner XML content and SOAP faults.
type embedBody struct {
	XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Body"`

	Content []byte      `xml:",innerxml"`
	Fault   *soap.Fault `xml:",omitempty"`
}
