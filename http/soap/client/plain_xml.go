package client

type PlainXml struct {
	Value []byte `xml:",innerxml"`
}
