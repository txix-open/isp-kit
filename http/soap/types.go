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

	Content interface{} `xml:",omitempty"`

	// faultOccurred indicates whether the XML body included a fault;
	// we cannot simply store SOAPFault as a pointer to indicate this, since
	// fault is initialized to non-nil with user-provided detail type.
	faultOccurred bool
	Fault         *Fault `xml:",omitempty"`
}

// UnmarshalXML copied from https://github.com/hooklift/gowsdl/blob/master/soap/soap.go
func (b *Body) UnmarshalXML(d *xml.Decoder, _ xml.StartElement) error {
	var (
		token    xml.Token
		err      error
		consumed bool
	)

Loop:
	for {
		if token, err = d.Token(); err != nil {
			return err
		}

		if token == nil {
			break
		}

		switch se := token.(type) {
		case xml.StartElement:
			if consumed {
				return xml.UnmarshalError("Found multiple elements inside SOAP body; not wrapped-document/literal WS-I compliant")
			} else if se.Name.Space == "http://schemas.xmlsoap.org/soap/envelope/" && se.Name.Local == "Fault" {
				b.Content = nil

				b.faultOccurred = true
				err = d.DecodeElement(b.Fault, &se)
				if err != nil {
					return err
				}

				consumed = true
			} else {
				if err = d.DecodeElement(b.Content, &se); err != nil {
					return err
				}

				consumed = true
			}
		case xml.EndElement:
			break Loop
		}
	}

	return nil
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
