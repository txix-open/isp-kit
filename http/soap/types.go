package soap

import (
	"encoding/xml"
	"net/http"

	"github.com/pkg/errors"
)

const (
	// ContentType is the standard SOAP content type.
	ContentType = `text/xml; charset="utf-8"`
)

// Envelope represents a SOAP envelope containing a header and body.
// It is the root element of a SOAP message.
type Envelope struct {
	XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Envelope"`

	Header *Header
	Body   Body
}

// Header represents a SOAP header containing optional metadata.
type Header struct {
	XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Header"`

	Items []any `xml:",omitempty"`
}

// Body represents a SOAP body containing the message content or fault.
// It supports both regular content and SOAP faults, with WS-I compliance validation.
type Body struct {
	XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Body"`

	Content any `xml:",omitempty"`

	// faultOccurred indicates whether the XML body included a fault;
	// we cannot simply store SOAPFault as a pointer to indicate this, since
	// fault is initialized to non-nil with user-provided detail type.
	faultOccurred bool
	Fault         *Fault `xml:",omitempty"`
}

// UnmarshalXML decodes a SOAP body from XML, handling both content and faults.
// It enforces WS-I compliance by rejecting multiple elements inside the SOAP body.
func (b *Body) UnmarshalXML(d *xml.Decoder, _ xml.StartElement) error {
	var (
		token    xml.Token
		err      error
		consumed bool
	)

Loop:
	for {
		token, err = d.Token()
		if err != nil {
			return err
		}

		if token == nil {
			break
		}

		switch se := token.(type) {
		case xml.StartElement:
			switch {
			case consumed:
				return xml.UnmarshalError("Found multiple elements inside SOAP body; not wrapped-document/literal WS-I compliant")
			case se.Name.Space == "http://schemas.xmlsoap.org/soap/envelope/" && se.Name.Local == "Fault":
				b.Content = nil

				b.faultOccurred = true
				err = d.DecodeElement(b.Fault, &se)
				if err != nil {
					return err
				}

				consumed = true
			default:
				if err = d.DecodeElement(b.Content, &se); err != nil { //nolint:noinlineerr
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

// Fault represents a SOAP fault containing error information.
// It follows the SOAP 1.1 fault structure with code, string, actor, and optional detail.
// nolint:errname
type Fault struct {
	XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault"`

	Code   string `xml:"faultcode,omitempty"`
	String string `xml:"faultstring,omitempty"`
	Actor  string `xml:"faultactor,omitempty"`
	Detail any    `xml:"detail,omitempty"`
}

// Error returns the fault string as the error message.
func (f Fault) Error() string {
	return f.String
}

// WriteError writes the fault as a SOAP XML response to the http.ResponseWriter.
// It sets the Content-Type to text/xml and the HTTP status code to 500.
func (f Fault) WriteError(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", ContentType)
	w.WriteHeader(http.StatusInternalServerError)
	env := Envelope{Body: Body{Fault: &f}}
	err := xml.NewEncoder(w).Encode(env)
	if err != nil {
		return errors.WithMessage(err, "encode xml")
	}
	return nil
}
