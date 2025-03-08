package client_test

import (
	"context"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/txix-open/isp-kit/http/endpoint/log_middleware"

	"github.com/txix-open/isp-kit/http/httpclix"
	"github.com/txix-open/isp-kit/http/soap"
	"github.com/txix-open/isp-kit/http/soap/client"
	"github.com/txix-open/isp-kit/test"
)

type Book struct {
	XMLName xml.Name `xml:"Book"`

	Id   int
	Name string
}

func TestClient_Invoke(t *testing.T) {
	test, require := test.New(t)
	handler := func(ctx context.Context, book Book) (*Book, error) {
		return &book, nil
	}
	wrapper := soap.DefaultWrapper(test.Logger(), log_middleware.Log(test.Logger(), log_middleware.WithLogBody(true)))
	mux := soap.NewActionMux().Handle("test", wrapper.Endpoint(handler))
	srv := httptest.NewServer(mux)
	cli := client.New(httpclix.Default())

	req := Book{Id: 123, Name: "test"}
	res := Book{}
	resp, err := cli.Invoke(context.Background(), srv.URL, "test", nil, req)
	require.NoError(err)
	require.True(resp.Http.IsSuccess())
	require.NoError(resp.UnmarshalPayload(&res))
	require.EqualValues(req.Id, res.Id)
	require.EqualValues(req.Name, res.Name)

	plainReq := client.PlainXml{Value: []byte("<Book><Id>123</Id><Name>test</Name></Book>")}
	plainResp := client.PlainXml{}
	resp, err = cli.Invoke(context.Background(), srv.URL, "test", nil, plainReq)
	require.NoError(err)
	require.True(resp.Http.IsSuccess())
	require.NoError(resp.UnmarshalPayload(&plainResp))
	require.EqualValues(plainReq, plainResp)

	req = Book{Id: 123, Name: "test"}
	res = Book{}
	resp, err = cli.Invoke(context.Background(), srv.URL, "unknown_action", nil, req)
	require.NoError(err)
	require.EqualValues(http.StatusInternalServerError, resp.Http.StatusCode())
	fault, err := resp.Fault()
	require.NoError(err)
	require.EqualValues("Client", fault.Code)
}
