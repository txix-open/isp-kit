package soap

import (
	"encoding/xml"
	"net/http"

	"github.com/pkg/errors"
)

const (
	ContentType = `text/xml; charset="utf-8"`
)

type Envelope struct {
	XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Envelope"`

	Header *Header
	Body   Body
}

type Header struct {
	XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Header"`

	Items []any `xml:",omitempty"`
}

type Body struct {
	XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Body"`

	Content []byte `xml:",innerxml"`
	Fault   *Fault `xml:",omitempty"`
}

type Fault struct {
	XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault"`

	Code   string `xml:"faultcode,omitempty"`
	String string `xml:"faultstring,omitempty"`
	Actor  string `xml:"faultactor,omitempty"`
	Detail any    `xml:"detail,omitempty"`
}

func (f Fault) Error() string {
	return f.String
}

func (f Fault) WriteError(w http.ResponseWriter) error {
	w.Header().Set("content-type", ContentType)
	w.WriteHeader(http.StatusInternalServerError)
	env := Envelope{Body: Body{Fault: &f}}
	err := xml.NewEncoder(w).Encode(env)
	if err != nil {
		return errors.WithMessage(err, "encode xml")
	}
	return nil
}
