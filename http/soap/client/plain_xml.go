package client

// PlainXml represents raw XML content for SOAP requests or responses.
// It is used when the SOAP body contains non-structured XML data.
// The Value field captures the inner XML without envelope wrapping.
type PlainXml struct {
	Value []byte `xml:",innerxml"`
}
