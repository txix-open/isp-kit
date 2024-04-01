package soap_test

import (
	"context"
	"encoding/xml"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/txix-open/isp-kit/http/httpcli"
	"github.com/txix-open/isp-kit/http/soap"
	"github.com/txix-open/isp-kit/log"
)

type Req struct {
	XMLName xml.Name `xml:"http://xmlns.example.com/sudir/connector Req"`

	EntryItem EntryItem
}

type EntryItem struct {
	EntryName string
}

func TestNamespace(t *testing.T) {
	require := require.New(t)
	body := `<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:con="http://xmlns.example.com/sudir/connector">
    <soapenv:Header/>
    
    <soapenv:Body>
        <con:Req>
            <con:EntryItem>
                <con:EntryName>Test</con:EntryName>
            </con:EntryItem>
        </con:Req>
    </soapenv:Body>
</soapenv:Envelope>`
	logger, err := log.New()
	require.NoError(err)
	wrapper := soap.DefaultWrapper(logger)
	handler := wrapper.Endpoint(func(ctx context.Context, req Req) {
		require.EqualValues("Test", req.EntryItem.EntryName)
	})
	httpHandler := soap.NewActionMux().Handle("Endpoint", handler)
	srv := httptest.NewServer(httpHandler)
	resp, err := httpcli.New().Post(srv.URL).
		Header(soap.ActionHeader, "Endpoint").
		Header("content-type", soap.ContentType).
		RequestBody([]byte(body)).
		Do(context.Background())
	require.NoError(err)
	require.True(resp.IsSuccess())

}
